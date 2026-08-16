package input

import (
	"reflect"
	"testing"

	"github.com/IBM/sarama"
)

func TestKafkaInputConfigParsing(t *testing.T) {
	tests := []struct {
		name     string
		config   map[any]any
		expected KafkaInputConfig
		panics   bool
	}{
		{
			name: "full valid kafka config with topics",
			config: map[any]any{
				"brokers":               []string{"localhost:9092", "localhost:9093"},
				"topics":                []string{"test-topic", "other-topic"},
				"group_id":              "logflow-group",
				"codec":                 "json",
				"worker":                4,
				"decorate_events":       false,
				"discard_on_error":      true,
				"messages_queue_length": 64,
				"auto_offset_earliest":  true,
				"auto_commit_interval":  30,
				"client_id":             "logflow-test",
			},
			expected: KafkaInputConfig{
				brokers:             []string{"localhost:9092", "localhost:9093"},
				topics:              []string{"test-topic", "other-topic"},
				topicPattern:        "",
				codec:               "json",
				groupID:             "logflow-group",
				decorateEvents:      false,
				worker:              4,
				messagesQueueLength: 64,
				discardOnError:      true,
			},
			panics: false,
		},
		{
			name: "minimal valid kafka config",
			config: map[any]any{
				"brokers":  []string{"localhost:9092"},
				"topics":   []string{"test-topic"},
				"group_id": "default-group",
			},
			expected: KafkaInputConfig{
				brokers:             []string{"localhost:9092"},
				topics:              []string{"test-topic"},
				codec:               "plain",
				groupID:             "default-group",
				decorateEvents:      true,
				worker:              1,
				messagesQueueLength: 8,
				discardOnError:      false,
			},
			panics: false,
		},
		{
			name: "group_id wrong type should panic",
			config: map[any]any{
				"brokers":  []string{"localhost:9092"},
				"topics":   []string{"t"},
				"group_id": 123,
			},
			panics: true,
		},
		{
			name: "topic_pattern wrong type should panic",
			config: map[any]any{
				"brokers":       []string{"localhost:9092"},
				"topic_pattern": 999,
				"group_id":      "g",
			},
			panics: true,
		},
		{
			name: "version wrong type should panic",
			config: map[any]any{
				"brokers":  []string{"localhost:9092"},
				"topics":   []string{"t"},
				"group_id": "g",
				"version":  35,
			},
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("newKafkaInputConfig() should have panicked")
					}
				}()
				newKafkaInputConfig(tt.config)
			} else {
				cfg := newKafkaInputConfig(tt.config)
				if cfg == nil {
					t.Fatal("newKafkaInputConfig() returned nil")
				}
				if cfg.clientConfig == nil {
					t.Error("clientConfig should be initialized")
				}
				if !equalKafkaInputConfig(*cfg, tt.expected) {
					t.Errorf("newKafkaInputConfig() = %+v, want %+v", cfg, tt.expected)
				}
				// Assert additional expected clientConfig fields
				assertClientConfigFields(t, tt.config, cfg.clientConfig)
			}
		})
	}
}

// assertClientConfigFields 验证 sarama.Config 中与配置键对应的字段
func assertClientConfigFields(t *testing.T, config map[any]any, cc *sarama.Config) {
	t.Helper()

	// client_id 默认 "logflow"
	if v, ok := config["client_id"]; ok {
		if cc.ClientID != v.(string) {
			t.Errorf("clientConfig.ClientID = %q, want %q", cc.ClientID, v.(string))
		}
	} else {
		if cc.ClientID != "logflow" {
			t.Errorf("clientConfig.ClientID = %q, want default %q", cc.ClientID, "logflow")
		}
	}

	// from_beginning: true -> OffsetOldest, false/absent -> OffsetNewest
	if v, ok := config["from_beginning"]; ok {
		if v.(bool) {
			if cc.Consumer.Offsets.Initial != sarama.OffsetOldest {
				t.Errorf("clientConfig.Consumer.Offsets.Initial = %d, want OffsetOldest (%d)",
					cc.Consumer.Offsets.Initial, sarama.OffsetOldest)
			}
		} else {
			if cc.Consumer.Offsets.Initial != sarama.OffsetNewest {
				t.Errorf("clientConfig.Consumer.Offsets.Initial = %d, want OffsetNewest (%d)",
					cc.Consumer.Offsets.Initial, sarama.OffsetNewest)
			}
		}
	}

	// enable_auto_commit
	if v, ok := config["enable_auto_commit"]; ok {
		if cc.Consumer.Offsets.AutoCommit.Enable != v.(bool) {
			t.Errorf("clientConfig.Consumer.Offsets.AutoCommit.Enable = %v, want %v",
				cc.Consumer.Offsets.AutoCommit.Enable, v.(bool))
		}
	}

	// auto_commit_interval (秒)
	if v, ok := config["auto_commit_interval"]; ok {
		want := int64(v.(int))
		got := int64(cc.Consumer.Offsets.AutoCommit.Interval.Seconds())
		if got != want {
			t.Errorf("clientConfig.Consumer.Offsets.AutoCommit.Interval = %ds, want %ds", got, want)
		}
	}

	// sasl
	if v, ok := config["sasl_enable"]; ok && v.(bool) {
		if !cc.Net.SASL.Enable {
			t.Error("clientConfig.Net.SASL.Enable should be true")
		}
		if cc.Net.SASL.User != config["sasl_username"].(string) {
			t.Errorf("SASL.User = %q, want %q", cc.Net.SASL.User, config["sasl_username"].(string))
		}
		if cc.Net.SASL.Password != config["sasl_password"].(string) {
			t.Errorf("SASL.Password = %q, want %q", cc.Net.SASL.Password, config["sasl_password"].(string))
		}
		if string(cc.Net.SASL.Mechanism) != config["sasl_mechanism"].(string) {
			t.Errorf("SASL.Mechanism = %q, want %q", cc.Net.SASL.Mechanism, config["sasl_mechanism"].(string))
		}
	}
}

// equalKafkaInputConfig 比较两个 KafkaInputConfig 的非 clientConfig 字段
func equalKafkaInputConfig(a, b KafkaInputConfig) bool {
	if a.codec != b.codec ||
		a.groupID != b.groupID ||
		a.topicPattern != b.topicPattern ||
		a.decorateEvents != b.decorateEvents ||
		a.discardOnError != b.discardOnError ||
		a.worker != b.worker ||
		a.messagesQueueLength != b.messagesQueueLength {
		return false
	}
	if !reflect.DeepEqual(a.brokers, b.brokers) {
		return false
	}
	if !reflect.DeepEqual(a.topics, b.topics) {
		return false
	}
	return true
}
