package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
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
	// config        map[any]any
	index         string
	indexFG       field_getter.FieldGetter
	hosts         []string
	user          string
	password      string
	apiKey        string // Base64 编码的 API Key 字符串（由 id:api_key 组成后 Base64），优先级高于 user/password
	ssl           bool
	cacert        string
	version       int8
	sniff         bool
	sniffInterval int
	maxRetries    int // 最大重试次数，默认 3 次

	client  *elasticsearch.Client
	bufPool sync.Pool
}

// newElasticsearchConfig 解析配置，返回 ElasticsearchOutput（不含 client）和 elasticsearch.Config。
// 不涉及文件 IO 和网络连接，可在单元测试中直接调用。
func newElasticsearchConfig(conf map[any]any) (*ElasticsearchOutput, elasticsearch.Config) {
	if conf == nil {
		logger.Fatal("elasticsearch output: config is nil")
	}
	e := &ElasticsearchOutput{
		// config:     conf,
		ssl:        false,
		version:    7,
		sniff:      false,
		maxRetries: 3,
		bufPool: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
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
		if v, ok := conf["cacert"]; ok {
			if e.cacert, ok = utils.ParseToStr(v); !ok {
				logger.Fatal("elasticsearch output: cacert config must be a string")
			}
		} else {
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

	clientConfig := elasticsearch.Config{
		Addresses:            e.hosts,
		Username:             e.user,
		Password:             e.password,
		APIKey:               e.apiKey,
		DisableRetry:         e.maxRetries == 0,
		MaxRetries:           e.maxRetries,
		DiscoverNodesOnStart: e.sniff,
		Transport: &http.Transport{
			ResponseHeaderTimeout: 10 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
		},
	}

	if e.sniff {
		clientConfig.DiscoverNodesInterval = time.Duration(e.sniffInterval) * time.Second
	}

	return e, clientConfig
}

func newElasticsearchOutput(config map[any]any) model.Output {
	e, clientConfig := newElasticsearchConfig(config)

	var err error
	if e.ssl {
		// 根据路径加载证书文件
		if clientConfig.CACert, err = os.ReadFile(e.cacert); err != nil {
			logger.Fatal("elasticsearch output: Failed to read Elasticsearch CACert file:"+e.cacert, zap.Error(err))
		}
	}

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

	return e
}

func (e *ElasticsearchOutput) Handle(event *event.Event) error {
	// 根据index格式和消息，生成需写入的index，获取失败则默认为字面量
	var index = e.index
	indexAny, err := e.indexFG.GetField(event)
	if err != nil {
		logger.Error("elasticsearch output: Failed to get index from event", zap.Error(err))
	} else if indexStr, ok := indexAny.(string); ok {
		index = indexStr
	}

	// 将 map 序列化为 JSON（复用 sync.Pool 中的 bytes.Buffer）
	buf := e.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer e.bufPool.Put(buf)

	if err = json.NewEncoder(buf).Encode(event.Data); err != nil {
		logger.Error("elasticsearch output: Failed to encode event data to JSON", zap.Error(err))
		return err
	}

	// TODO: 写入失败后将数据存入失败队列，不断重试。（待定：持久化到磁盘，避免重启导致数据丢失）
	// 写入事件到 Elasticsearch
	res, err := e.client.Index(index, buf)
	if err != nil {
		logger.Error("elasticsearch output: Failed to index event to Elasticsearch", zap.Error(err))
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		logger.Error("elasticsearch output: Failed to index event to Elasticsearch",
			zap.String("resp", res.String()))
		return fmt.Errorf("elasticsearch index failed: %s", res.Status())
	}
	logger.Debug("elasticsearch output: Indexed event to Elasticsearch", zap.String("resp", res.String()))
	return nil
}

func (e *ElasticsearchOutput) Close() {
	// 客户端不是长连接，不需要关闭
	logger.Info("elasticsearch output Close ok")
}
