package output

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Shadowmaple/logflow/field/field_getter"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/utils"
	"github.com/Shadowmaple/logflow/model"

	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/zap"
)

func init() {
	Register("elasticsearch", newElasticsearchOutput)
}

type ElasticsearchOutput struct {
	index         string
	indexFG       field_getter.FieldGetter
	hosts         []string
	user          string
	password      string
	apiKey        string // Base64 编码的 API Key 字符串（由 id:api_key 组成后 Base64），优先级高于 user/password
	ssl           bool
	cacert        string // CA 证书路径
	skipSSLVerify bool   // 是否跳过 SSL 验证，默认 false，生产环境不建议使用
	version       int8
	sniff         bool
	sniffInterval int
	maxRetries    int           // 最大重试次数，默认 3 次
	bulkCount     int           // 触发批量提交的消息条数，默认 5000
	bulkSize      int           // 触发批量提交的字节阈值，默认 15MB
	flushInterval time.Duration // 定时提交间隔，默认 30s，避免数据量过少时长时间不提交

	client    *elasticsearch.Client
	bulkMu    sync.Mutex // 保护 bulkBuf/bulkCur 在 Handle、定时 flush goroutine 与 Close 间的并发访问
	bulkBuf   *bytes.Buffer
	bulkCur   int
	metaCache map[string][]byte // index -> action 行字节缓存，避免重复 marshal
	flushStop chan struct{}     // 通知定时 flush goroutine 退出
	flusherWg sync.WaitGroup    // 等待定时 flush goroutine 结束
}

// newElasticsearchConfig 解析配置，返回 ElasticsearchOutput（不含 client）和 elasticsearch.Config。
// 不涉及文件 IO 和网络连接，可在单元测试中直接调用。
func parseElasticsearchConfig(conf map[any]any) *ElasticsearchOutput {
	if conf == nil {
		logger.Fatal("elasticsearch output: config is nil")
	}
	e := &ElasticsearchOutput{
		ssl:           false,
		version:       7,
		sniff:         false,
		maxRetries:    3,
		bulkCount:     5000,
		bulkSize:      15 * 1024 * 1024, // 15MB
		flushInterval: 30 * time.Second,
		bulkBuf:       new(bytes.Buffer),
		metaCache:     make(map[string][]byte),
	}
	if v, ok := conf["index"]; ok {
		if e.index, ok = utils.ParseToStr(v); !ok {
			logger.Fatal("elasticsearch output: index must be a string")
		}
	} else {
		logger.Fatal("elasticsearch output: index config is required")
	}
	e.indexFG = field_getter.GetFieldGetter2(e.index)

	if v, ok := conf["hosts"]; ok {
		e.hosts, ok = utils.ParseToStrList(v)
		if !ok {
			logger.Fatal("elasticsearch output hosts config must be a list of strings")
		}
	} else {
		logger.Fatal("elasticsearch output hosts config is required")
	}
	if v, ok := conf["user"]; ok {
		if e.user, ok = utils.ParseToStr(v); !ok {
			logger.Fatal("elasticsearch output: user must be a string")
		}
	}
	if v, ok := conf["password"]; ok {
		if e.password, ok = utils.ParseToStr(v); !ok {
			logger.Fatal("elasticsearch output: password must be a string")
		}
	}
	if v, ok := conf["ssl"]; ok {
		e.ssl, ok = utils.ParseToBool(v)
		if !ok {
			logger.Fatal("elasticsearch output: ssl config must be a boolean value")
		}
	}
	if e.ssl {
		if v, ok := conf["skip_ssl_verification"]; ok {
			e.skipSSLVerify, ok = utils.ParseToBool(v)
			if !ok {
				logger.Fatal("elasticsearch output: skip_ssl_verification config must be a boolean value")
			}
		}
		if v, ok := conf["cacert"]; ok {
			if e.cacert, ok = utils.ParseToStr(v); !ok {
				logger.Fatal("elasticsearch output: cacert config must be a string")
			}
		}
		if !e.skipSSLVerify && e.cacert == "" {
			logger.Fatal("elasticsearch output: cacert config is required")
		}
	}
	if v, ok := conf["api_key"]; ok {
		if e.apiKey, ok = utils.ParseToStr(v); !ok {
			logger.Fatal("elasticsearch output: api_key must be a string")
		}
	}
	if v, ok := conf["sniff"]; ok {
		if e.sniff, ok = utils.ParseToBool(v); !ok {
			logger.Fatal("elasticsearch output: sniff config must be a boolean value")
		}
	}
	if e.sniff {
		if v, ok := conf["sniff_interval"]; ok {
			if e.sniffInterval, ok = utils.ParseToInt(v); !ok || e.sniffInterval < 0 {
				logger.Fatal("elasticsearch output: sniff_interval config must be greater than or equal to 0")
			}
		} else {
			e.sniffInterval = 30
		}
	}
	if v, ok := conf["max_retries"]; ok {
		if e.maxRetries, ok = utils.ParseToInt(v); !ok || e.maxRetries < 0 {
			logger.Fatal("elasticsearch output: max_retries config must be greater than or equal to 0")
		}
	}
	// 批量提交触发条件：bulk_count 条消息 或 bulk_size MB 数据，任一满足即提交
	if v, ok := conf["bulk_count"]; ok {
		if e.bulkCount, ok = utils.ParseToInt(v); !ok || e.bulkCount <= 0 {
			logger.Fatal("elasticsearch output: bulk_count must be a positive integer")
		}
	}
	if v, ok := conf["bulk_size"]; ok {
		bulkSizeMB, ok := utils.ParseToInt(v)
		if !ok || bulkSizeMB <= 0 {
			logger.Fatal("elasticsearch output: bulk_size must be a positive integer (MB)")
		}
		e.bulkSize = bulkSizeMB * 1024 * 1024
	}
	// 定时提交间隔（秒）：即使未达到 bulk_count/bulk_size，也每隔该时间提交一次缓冲区
	if v, ok := conf["flush_interval"]; ok {
		fi, ok := utils.ParseToInt(v)
		if !ok || fi <= 0 {
			logger.Fatal("elasticsearch output: flush_interval must be a positive integer (seconds)")
		}
		e.flushInterval = time.Duration(fi) * time.Second
	}

	return e
}

func newElasticsearchOutput(config map[any]any) model.Output {
	// 解析配置并创建 ElasticsearchOutput 实例
	e := parseElasticsearchConfig(config)

	transport := &http.Transport{
		ResponseHeaderTimeout: 10 * time.Second,
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
	}
	if e.ssl {
		if e.skipSSLVerify {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		} else {
			caPool, err := utils.LoadCACert(e.cacert)
			if err != nil {
				logger.Fatal("elasticsearch output: load ca cert failed", zap.Error(err))
			}
			transport.TLSClientConfig = &tls.Config{RootCAs: caPool}
		}
	}

	clientConfig := elasticsearch.Config{
		Addresses:            e.hosts,
		Username:             e.user,
		Password:             e.password,
		APIKey:               e.apiKey,
		DisableRetry:         e.maxRetries == 0,
		MaxRetries:           e.maxRetries,
		DiscoverNodesOnStart: e.sniff,
		Transport:            transport,
	}

	if e.sniff {
		clientConfig.DiscoverNodesInterval = time.Duration(e.sniffInterval) * time.Second
	}

	var err error
	e.client, err = elasticsearch.NewClient(clientConfig)
	if err != nil {
		logger.Fatal("elasticsearch output: Failed to create Elasticsearch client", zap.Error(err))
	}

	// 检查 Elasticsearch 连接是否成功
	res, err := e.client.Info()
	if err != nil {
		logger.Fatal("elasticsearch output: Failed to get Elasticsearch info", zap.Error(err))
	}
	defer res.Body.Close()
	if res.IsError() {
		logger.Fatal("elasticsearch output: Failed to get Elasticsearch info", zap.String("body", res.String()))
	}
	logger.Info("elasticsearch output: Connected to Elasticsearch", zap.String("body", res.String()))

	// 启动定时 flush goroutine，按 flush_interval 周期性提交缓冲区
	e.flushStop = make(chan struct{})
	e.flusherWg.Add(1)
	go e.runFlusher()

	return e
}

// Handle 将事件追加到批量缓冲区，当条数（bulkCount）或字节数（bulkSize）达到阈值时触发批量提交。
func (e *ElasticsearchOutput) Handle(event *event.Event) error {
	// 根据index格式和消息，生成需写入的index，获取失败则默认为字面量
	var index = e.index
	indexAny, err := e.indexFG.GetField(event)
	if err != nil {
		logger.Error("elasticsearch output: Failed to get index from event", zap.Error(err))
	} else if indexStr, ok := indexAny.(string); ok {
		index = indexStr
	}

	e.bulkMu.Lock()
	defer e.bulkMu.Unlock()

	// 写入 action 行：{"index":{"_index":"<index>"}}
	meta, err := e.metaLine(index)
	if err != nil {
		logger.Error("elasticsearch output: marshal bulk meta failed", zap.Error(err))
		return err
	}
	metaStart := e.bulkBuf.Len()
	e.bulkBuf.Write(meta)

	// 写入文档行。json.Encoder.Encode 先在内部完成序列化，成功后才写入底层 writer，
	// 失败时 bulkBuf 不会被改动（仅残留上面的 action 行，需回滚）。
	if err = json.NewEncoder(e.bulkBuf).Encode(event.Data); err != nil {
		e.bulkBuf.Truncate(metaStart)
		logger.Error("elasticsearch output: Failed to encode event data to JSON", zap.Error(err))
		return err
	}

	e.bulkCur++
	// 达到任一阈值即提交
	if e.bulkCur >= e.bulkCount || e.bulkBuf.Len() >= e.bulkSize {
		return e.flush()
	}
	return nil
}

// flush 提交当前缓冲区中的批量请求。调用方必须持有 e.bulkMu。
func (e *ElasticsearchOutput) flush() error {
	if e.bulkCur == 0 {
		return nil
	}
	res, err := e.client.Bulk(bytes.NewReader(e.bulkBuf.Bytes()))
	if err != nil {
		count := e.bulkCur
		e.resetBulk()
		// TODO: 写入失败后将数据存入失败队列，不断重试。（待定：持久化到磁盘，避免重启导致数据丢失）
		logger.Error("elasticsearch output: bulk index request failed",
			zap.Error(err), zap.Int("dropped", count))
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		count := e.bulkCur
		status := res.Status()
		logger.Error("elasticsearch output: bulk index request failed",
			zap.Int("dropped", count), zap.String("status", status), zap.String("resp", res.String()))
		e.resetBulk()
		return fmt.Errorf("elasticsearch bulk index failed: %s", status)
	}
	// 批量响应体中可能存在单项失败（HTTP 200 但 errors=true），检查并记录
	var result struct {
		Errors bool `json:"errors"`
		Took   int  `json:"took"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		logger.Error("elasticsearch output: decode bulk response failed", zap.Error(err))
	} else if result.Errors {
		logger.Error("elasticsearch output: bulk request had item-level errors",
			zap.Int("count", e.bulkCur), zap.Int("took", result.Took))
	}
	logger.Debug("elasticsearch output: bulk indexed ok",
		zap.Int("count", e.bulkCur), zap.Int("bytes", e.bulkBuf.Len()))
	e.resetBulk()
	return nil
}

// resetBulk 清空批量缓冲区。调用方必须持有 e.bulkMu。
func (e *ElasticsearchOutput) resetBulk() {
	e.bulkBuf.Reset()
	e.bulkCur = 0
}

// metaLine 返回指定 index 对应的 bulk action 行（含换行）。
// 结果按 index 缓存，避免每条消息都重新 marshal。调用方必须持有 e.bulkMu。
func (e *ElasticsearchOutput) metaLine(index string) ([]byte, error) {
	if line, ok := e.metaCache[index]; ok {
		return line, nil
	}
	meta := map[string]map[string]string{"index": {"_index": index}}
	line, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	line = append(line, '\n')
	e.metaCache[index] = line
	return line, nil
}

// runFlusher 按固定间隔触发批量提交，避免数据量过少时长时间停留在缓冲区。
func (e *ElasticsearchOutput) runFlusher() {
	defer e.flusherWg.Done()
	ticker := time.NewTicker(e.flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-e.flushStop:
			return
		case <-ticker.C:
			e.bulkMu.Lock()
			if e.bulkCur > 0 {
				if err := e.flush(); err != nil {
					logger.Error("elasticsearch output: timed flush failed", zap.Error(err))
				}
			}
			e.bulkMu.Unlock()
		}
	}
}

func (e *ElasticsearchOutput) Close() {
	// 先停止定时 flush goroutine，避免与最终的 flush 竞争
	if e.flushStop != nil {
		close(e.flushStop)
		e.flusherWg.Wait()
	}
	// 关闭前刷新剩余的缓冲事件，避免数据丢失
	e.bulkMu.Lock()
	if e.bulkCur > 0 {
		if err := e.flush(); err != nil {
			logger.Error("elasticsearch output: flush on close failed", zap.Error(err))
		}
	}
	e.bulkMu.Unlock()
	logger.Info("elasticsearch output Close ok")
}
