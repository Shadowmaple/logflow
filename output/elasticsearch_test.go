package output

import (
	"reflect"
	"testing"
	"time"

	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 让 logger.Fatal 在测试中触发 panic 而非 os.Exit，便于 recover 捕获
func init() {
	logger.Logger = zap.New(zapcore.NewNopCore(), zap.WithFatalHook(zapcore.WriteThenPanic))
}

// assertPanics 断言 fn 执行时会 panic
func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, got none")
		}
	}()
	fn()
}

func TestNewElasticsearchConfig(t *testing.T) {
	tests := []struct {
		name             string
		config           map[any]any
		wantOutput       ElasticsearchOutput
		wantClientConfig elasticsearch.Config
		panics           bool
	}{
		{
			name: "full valid config",
			config: map[any]any{
				"index":          "logflow-%{+YYYY.MM.dd}",
				"hosts":          []string{"http://localhost:9200", "http://localhost:9201"},
				"user":           "elastic",
				"password":       "secret",
				"api_key":        "aGVsbG86d29ybGQ=",
				"ssl":            true,
				"cacert":         "/etc/ssl/certs/ca.crt",
				"sniff":          true,
				"sniff_interval": 60,
				"max_retries":    5,
			},
			wantOutput: ElasticsearchOutput{
				index:         "logflow-%{+YYYY.MM.dd}",
				hosts:         []string{"http://localhost:9200", "http://localhost:9201"},
				user:          "elastic",
				password:      "secret",
				apiKey:        "aGVsbG86d29ybGQ=",
				ssl:           true,
				cacert:        "/etc/ssl/certs/ca.crt",
				sniff:         true,
				sniffInterval: 60,
				maxRetries:    5,
			},
			wantClientConfig: elasticsearch.Config{
				Addresses:             []string{"http://localhost:9200", "http://localhost:9201"},
				Username:              "elastic",
				Password:              "secret",
				APIKey:                "aGVsbG86d29ybGQ=",
				DisableRetry:          false,
				MaxRetries:            5,
				DiscoverNodesOnStart:  true,
				DiscoverNodesInterval: 60 * time.Second,
			},
			panics: false,
		},
		{
			name: "minimal config with string hosts",
			config: map[any]any{
				"index": "logs",
				"hosts": "http://localhost:9200",
			},
			wantOutput: ElasticsearchOutput{
				index:         "logs",
				hosts:         []string{"http://localhost:9200"},
				user:          "",
				password:      "",
				apiKey:        "",
				ssl:           false,
				cacert:        "",
				sniff:         false,
				sniffInterval: 0,
				maxRetries:    3,
			},
			wantClientConfig: elasticsearch.Config{
				Addresses:             []string{"http://localhost:9200"},
				DisableRetry:          false,
				MaxRetries:            3,
				DiscoverNodesOnStart:  false,
				DiscoverNodesInterval: 0,
			},
			panics: false,
		},
		{
			name: "sniff enabled with default interval",
			config: map[any]any{
				"index": "logs",
				"hosts": []string{"http://localhost:9200"},
				"sniff": true,
			},
			wantOutput: ElasticsearchOutput{
				index:         "logs",
				hosts:         []string{"http://localhost:9200"},
				sniff:         true,
				sniffInterval: 30,
				maxRetries:    3,
			},
			wantClientConfig: elasticsearch.Config{
				Addresses:             []string{"http://localhost:9200"},
				MaxRetries:            3,
				DiscoverNodesOnStart:  true,
				DiscoverNodesInterval: 30 * time.Second,
			},
			panics: false,
		},
		{
			name: "max_retries zero disables retry",
			config: map[any]any{
				"index":       "logs",
				"hosts":       []string{"http://localhost:9200"},
				"max_retries": 0,
			},
			wantOutput: ElasticsearchOutput{
				index:      "logs",
				hosts:      []string{"http://localhost:9200"},
				maxRetries: 0,
			},
			wantClientConfig: elasticsearch.Config{
				Addresses:    []string{"http://localhost:9200"},
				DisableRetry: true,
				MaxRetries:   0,
			},
			panics: false,
		},
		{
			name:   "nil config should panic",
			config: nil,
			panics: true,
		},
		{
			name: "missing index should panic",
			config: map[any]any{
				"hosts": []string{"http://localhost:9200"},
			},
			panics: true,
		},
		{
			name: "missing hosts should panic",
			config: map[any]any{
				"index": "logs",
			},
			panics: true,
		},
		{
			name: "index not a string should panic",
			config: map[any]any{
				"index": 123,
				"hosts": []string{"http://localhost:9200"},
			},
			panics: true,
		},
		{
			name: "hosts invalid type should panic",
			config: map[any]any{
				"index": "logs",
				"hosts": 123,
			},
			panics: true,
		},
		{
			name: "ssl true without cacert should panic",
			config: map[any]any{
				"index": "logs",
				"hosts": []string{"http://localhost:9200"},
				"ssl":   true,
			},
			panics: true,
		},
		{
			name: "ssl invalid type should panic",
			config: map[any]any{
				"index": "logs",
				"hosts": []string{"http://localhost:9200"},
				"ssl":   "not-a-bool",
			},
			panics: true,
		},
		{
			name: "sniff invalid type should panic",
			config: map[any]any{
				"index": "logs",
				"hosts": []string{"http://localhost:9200"},
				"sniff": "yes",
			},
			panics: true,
		},
		{
			name: "sniff_interval negative should panic",
			config: map[any]any{
				"index":          "logs",
				"hosts":          []string{"http://localhost:9200"},
				"sniff":          true,
				"sniff_interval": -1,
			},
			panics: true,
		},
		{
			name: "max_retries negative should panic",
			config: map[any]any{
				"index":       "logs",
				"hosts":       []string{"http://localhost:9200"},
				"max_retries": -5,
			},
			panics: true,
		},
		{
			name: "user not a string should panic",
			config: map[any]any{
				"index": "logs",
				"hosts": []string{"http://localhost:9200"},
				"user":  123,
			},
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				assertPanics(t, func() { parseElasticsearchConfig(tt.config) })
				return
			}

			got := parseElasticsearchConfig(tt.config)
			if got == nil {
				t.Fatal("newElasticsearchConfig() returned nil output")
			}

			// 比较非 client 字段（不含 indexFG 和 bufPool，它们是指针/接口）
			if got.index != tt.wantOutput.index {
				t.Errorf("index = %q, want %q", got.index, tt.wantOutput.index)
			}
			if got.indexFG == nil {
				t.Error("indexFG should be initialized")
			}
			if !reflect.DeepEqual(got.hosts, tt.wantOutput.hosts) {
				t.Errorf("hosts = %v, want %v", got.hosts, tt.wantOutput.hosts)
			}
			if got.user != tt.wantOutput.user {
				t.Errorf("user = %q, want %q", got.user, tt.wantOutput.user)
			}
			if got.password != tt.wantOutput.password {
				t.Errorf("password = %q, want %q", got.password, tt.wantOutput.password)
			}
			if got.apiKey != tt.wantOutput.apiKey {
				t.Errorf("apiKey = %q, want %q", got.apiKey, tt.wantOutput.apiKey)
			}
			if got.ssl != tt.wantOutput.ssl {
				t.Errorf("ssl = %v, want %v", got.ssl, tt.wantOutput.ssl)
			}
			if got.cacert != tt.wantOutput.cacert {
				t.Errorf("cacert = %q, want %q", got.cacert, tt.wantOutput.cacert)
			}
			if got.sniff != tt.wantOutput.sniff {
				t.Errorf("sniff = %v, want %v", got.sniff, tt.wantOutput.sniff)
			}
			if got.sniffInterval != tt.wantOutput.sniffInterval {
				t.Errorf("sniffInterval = %d, want %d", got.sniffInterval, tt.wantOutput.sniffInterval)
			}
			if got.maxRetries != tt.wantOutput.maxRetries {
				t.Errorf("maxRetries = %d, want %d", got.maxRetries, tt.wantOutput.maxRetries)
			}
		})
	}
}
