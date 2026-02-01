package output

import (
	"bytes"
	"encoding/json"
	"os"
	"time"

	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/model"
	"github.com/Shadowmaple/logflow/internal/utils"
	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/zap"
)

type ElasticsearchOutput struct {
	config        map[string]any
	index         string
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

	client *elasticsearch.Client
}

func newElasticsearchOutput(config map[string]any) Output {
	if config == nil {
		logger.Fatal("Elasticsearch config is nil")
	}
	e := &ElasticsearchOutput{
		config:     config,
		ssl:        false,
		version:    7,
		sniff:      false,
		maxRetries: 3,
	}
	if v, ok := config["index"]; ok {
		e.index = utils.TrimStr(v.(string))
	}
	if v, ok := config["hosts"]; ok {
		e.hosts, ok = utils.ParseToStrList(v)
		if !ok {
			logger.Fatal("Elasticsearch hosts config must be a list of strings")
		}
	} else {
		logger.Fatal("Elasticsearch hosts config is required")
	}
	if v, ok := config["user"]; ok {
		e.user = utils.TrimStr(v.(string))
	}
	if v, ok := config["password"]; ok {
		e.password = utils.TrimStr(v.(string))
	}
	if v, ok := config["ssl"]; ok {
		e.ssl, ok = utils.ParseToBool(v)
		if !ok {
			logger.Fatal("Elasticsearch ssl config must be a boolean value")
		}
	}
	if e.ssl {
		if v, ok := config["cacert"]; ok {
			e.cacert = utils.TrimStr(v.(string))
		} else if e.ssl {
			logger.Fatal("Elasticsearch cacert config is required")
		}
	}
	if v, ok := config["api_key"]; ok {
		e.apiKey = utils.TrimStr(v.(string))
	}
	if v, ok := config["sniff"]; ok {
		e.sniff, ok = utils.ParseToBool(v)
		if !ok {
			logger.Fatal("Elasticsearch sniff config must be a boolean value")
		}
	}
	if e.sniff {
		if v, ok := config["sniff_interval"]; ok {
			e.sniffInterval, ok = utils.ParseToInt(v)
			if !ok || e.sniffInterval < 0 {
				logger.Fatal("Elasticsearch sniff_interval config must be greater than or equal to 0")
			}
		} else {
			e.sniffInterval = 30
		}
	}
	if v, ok := config["max_retries"]; ok {
		e.maxRetries, ok = utils.ParseToInt(v)
		if !ok || e.maxRetries < 0 {
			logger.Fatal("Elasticsearch max_retries config must be greater than or equal to 0")
		}
	}

	var err error
	clientConfig := elasticsearch.Config{
		Addresses:            e.hosts,
		Username:             e.user,
		Password:             e.password,
		APIKey:               e.apiKey,
		DisableRetry:         e.maxRetries == 0,
		MaxRetries:           e.maxRetries,
		DiscoverNodesOnStart: e.sniff,
	}

	if e.sniff {
		clientConfig.DiscoverNodesInterval = time.Duration(e.sniffInterval) * time.Second
	}

	if e.ssl {
		// 根据路径加载证书文件
		if clientConfig.CACert, err = os.ReadFile(e.cacert); err != nil {
			logger.Fatal("Failed to read Elasticsearch CACert file:"+e.cacert, zap.Error(err))
		}
	}

	e.client, err = elasticsearch.NewClient(clientConfig)
	if err != nil {
		logger.Fatal("Failed to create Elasticsearch client", zap.Error(err))
	}

	// 检查 Elasticsearch 连接是否成功
	// if res, err := e.client.Ping(); err != nil {
	// 	logger.Fatal("Failed to ping Elasticsearch", zap.Error(err))
	// } else if res.IsError() {
	// 	logger.Fatal("Failed to ping Elasticsearch", zap.String("body", res.String()))
	// }
	res, err := e.client.Info()
	if err != nil {
		logger.Fatal("Failed to get Elasticsearch info", zap.Error(err))
	}
	defer res.Body.Close()
	if res.IsError() {
		logger.Fatal("Failed to get Elasticsearch info", zap.String("body", res.String()))
	}
	logger.Info("Connected to Elasticsearch", zap.String("body", res.String()))

	return e
}

func (e *ElasticsearchOutput) Handle(event *model.Event) error {

	// TODO: 根据index格式和消息，生成需写入的index
	index := ""

	// 将 map 序列化为 JSON
	// TODO: 复用buf
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(event.Data); err != nil {
		logger.Error("Failed to encode event data to JSON", zap.Error(err))
		return err
	}

	// TODO: 写入失败后将数据存入失败队列，不断重试。（待定：持久化到磁盘，避免重启导致数据丢失）
	// 写入事件到 Elasticsearch
	res, err := e.client.Index(index, &buf)
	if err != nil {
		logger.Error("Failed to index event to Elasticsearch", zap.Error(err))
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		logger.Error("Failed to index event to Elasticsearch", zap.String("resp", res.String()))
		return err
	}
	logger.Debug("Indexed event to Elasticsearch", zap.String("resp", res.String()))
	return nil
}

func (e *ElasticsearchOutput) Close() {
	// 客户端不是长连接，不需要关闭
}
