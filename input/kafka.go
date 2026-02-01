package input

import (
	"context"
	"fmt"
	"time"

	"github.com/Shadowmaple/logflow/codec"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/model"
	"github.com/Shadowmaple/logflow/internal/utils"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type KafkaInputConfig struct {
	brokers             []string
	topics              []string
	topicPattern        string // 正则表达式topic，TODO：需手动实现，拉取所有->匹配
	codec               string
	groupID             string
	decorateEvents      bool
	worker              int // 并发消费数
	messagesQueueLength int
	discardOnError      bool // 遇到错误是否丢弃消息

	clientConfig *sarama.Config
}

type KafkaInput struct {
	config         map[string]any
	decorateEvents bool
	discardOnError bool // 遇到错误是否丢弃消息

	ch       chan *model.Event
	messages chan *sarama.ConsumerMessage
	stop     bool

	decoder codec.Decoder

	consumers      []*sarama.Consumer
	groupConsumers []*sarama.ConsumerGroup
}

// const (
// 	DEFAULT_KAFKA_QUEUE_LEN = 256
// )

func newKafkaInputConfig(conf map[string]any) *KafkaInputConfig {
	c := &KafkaInputConfig{
		codec:               "plain",
		worker:              1,
		decorateEvents:      true,
		messagesQueueLength: 12,
		discardOnError:      false,
	}

	if v, ok := conf["brokers"]; ok {
		c.brokers, ok = utils.ParseToStrList(v)
		if !ok {
			logger.Fatal("kafka input: parse config failed: invalid brokers")
		}
	}

	if v, ok := conf["topics"]; ok {
		c.topics = v.([]string)
	}
	if v, ok := conf["topic_pattern"]; ok {
		c.topicPattern = v.(string)
	}
	if c.topicPattern == "" && len(c.topics) == 0 {
		logger.Fatal("kafka input: parse config failed: topics or topic_pattern is required")
	}

	if v, ok := conf["group_id"]; ok {
		c.groupID = v.(string)
	} else {
		logger.Fatal("kafka input: parse config failed: group_id is required")
	}

	if v, ok := conf["worker"]; ok {
		c.worker, ok = utils.ParseToInt(v)
		if !ok || c.worker <= 0 {
			logger.Fatal("kafka input: parse config failed: invalid worker")
		}
	}

	if v, ok := conf["discard_on_error"]; ok {
		c.discardOnError, ok = utils.ParseToBool(v)
		if !ok {
			logger.Fatal("kafka input: parse config failed: invalid discard_on_error")
		}
	}

	if v, ok := conf["decorate_events"]; ok {
		c.decorateEvents, ok = utils.ParseToBool(v)
		if !ok {
			logger.Fatal("kafka input: parse config failed: invalid decorate_events")
		}
	}

	if v, ok := conf["messages_queue_length"]; ok {
		c.messagesQueueLength, ok = utils.ParseToInt(v)
		if !ok || c.messagesQueueLength <= 0 {
			logger.Fatal("kafka input: parse config failed: invalid messages_queue_length")
		}
	}

	// 初始化sarama客户端配置
	clientConfig := sarama.NewConfig()

	if v, ok := conf["from_beginning"]; ok {
		fromBeginning, ok := utils.ParseToBool(v)
		if !ok {
			logger.Fatal("kafka input: parse config failed: invalid from_beginning")
		}
		// 当消费者组第一次消费时，从哪个位置开始消费
		if fromBeginning {
			clientConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
		} else {
			clientConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
		}
	}

	// 是否自动提交偏移量
	if v, ok := conf["enable_auto_commit"]; ok {
		enableAutoCommit, ok := utils.ParseToBool(v)
		if !ok {
			logger.Fatal("kafka input: parse config failed: invalid enable_auto_commit")
		}
		clientConfig.Consumer.Offsets.AutoCommit.Enable = enableAutoCommit
	}

	// 提交偏移量的时间间隔
	if v, ok := conf["commit_interval"]; ok {
		commitInterval, ok := utils.ParseToInt(v)
		if !ok || commitInterval <= 0 {
			logger.Fatal("kafka input: parse config failed: invalid commit_interval")
		}
		clientConfig.Consumer.Offsets.AutoCommit.Interval = time.Duration(commitInterval) * time.Second
	}

	// 配置SASL认证
	if v, ok := conf["sasl_enable"]; ok {
		saslEnable, ok := utils.ParseToBool(v)
		if !ok {
			logger.Fatal("kafka input: parse config failed: invalid sasl_enable")
		}
		clientConfig.Net.SASL.Enable = saslEnable
		if saslEnable {
			clientConfig.Net.SASL.User = conf["sasl_username"].(string)
			clientConfig.Net.SASL.Password = conf["sasl_password"].(string)
			if clientConfig.Net.SASL.User == "" || clientConfig.Net.SASL.Password == "" {
				logger.Fatal("kafka input: parse config failed: invalid sasl_username and sasl_password")
			}
			clientConfig.Net.SASL.Mechanism = sarama.SASLMechanism(conf["sasl_mechanism"].(string))
		}
	}

	if v, ok := conf["version"]; ok {
		kafkaVersionn, err := sarama.ParseKafkaVersion(v.(string))
		if err != nil {
			logger.Fatal("kafka input: parse config failed: invalid version", zap.Error(err))
		}
		clientConfig.Version = kafkaVersionn
	}

	if v, ok := conf["client_id"]; ok {
		clientConfig.ClientID = v.(string)
	} else {
		clientConfig.ClientID = "logflow"
	}

	// 消费者超时设置
	// clientConfig.Consumer.Group.Session.Timeout = 30 * time.Second
	// clientConfig.Consumer.Group.Heartbeat.Interval = 10 * time.Second

	if err := clientConfig.Validate(); err != nil {
		logger.Fatal("kafka input: config validate failed", zap.Error(err))
	}

	c.clientConfig = clientConfig
	return c
}

func newKafkaInput(conf map[string]any) Input {
	// 解析config
	config := newKafkaInputConfig(conf)
	kafkaInput := &KafkaInput{
		config:         conf,
		decorateEvents: config.decorateEvents,
		discardOnError: config.discardOnError,
		ch:             make(chan *model.Event, config.messagesQueueLength),
		messages:       make(chan *sarama.ConsumerMessage, config.messagesQueueLength),
		groupConsumers: make([]*sarama.ConsumerGroup, config.worker),
		decoder:        codec.NewDecoder(config.codec),
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	for i := 0; i < config.worker; i++ {
		// 创建消费者组
		client, err := sarama.NewConsumerGroup(config.brokers, config.groupID, config.clientConfig)
		if err != nil {
			logger.Fatal("kafka input: create consumer group failed", zap.Error(err))
		}
		kafkaInput.groupConsumers[i] = &client
		go func() {
			for err := range client.Errors() {
				logger.Error("kafka input: client err:" + err.Error())
			}
		}()

		go func() {
			defer func() {
				if err := client.Close(); err != nil {
					logger.Error("kafka input: client close err", zap.Error(err))
				}
			}()

			// 启动消费
			for ctx.Err() == nil {
				// 开始消费，需要传入上下文、主题列表和消费者处理器
				if err := client.Consume(ctx, config.topics, kafkaInput); err != nil {
					// 如果是上下文取消错误，则正常退出
					if err == sarama.ErrClosedConsumerGroup {
						logger.Info("kafka input: consume closed by ctx", zap.Error(err))
						return
					}
					logger.Error("kafka input: consume process err, will retry", zap.Error(err))
					// 发生错误时等待一段时间再重试
					// TODO: 随机时间
					time.Sleep(5 * time.Second)
				}
			}
		}()
	}

	return kafkaInput
}

// receive events
func (ki *KafkaInput) Receive() <-chan *model.Event {
	return ki.ch
}

// receive events
func (ki *KafkaInput) ReceiveOne() *model.Event {
	if ki.stop {
		return nil
	}
	for !ki.stop {
		msg, ok := <-ki.messages
		if !ok {
			return nil
		}

		logger.Debug(
			fmt.Sprintf("kafka input: 主题: %s, 分区: %d, 偏移量: %d, 时间戳: %v, 消息内容: %s",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				msg.Timestamp,
				string(msg.Value),
			),
		)
		if len(msg.Value) == 0 {
			logger.Warn("kafka input: msg value is nil", zap.String("topic", msg.Topic))
			// continue
		}

		data, err := ki.decoder.Decode(msg.Value)
		if err != nil || data == nil {
			logger.Error("kafka input: decoded msg failed: "+string(msg.Value), zap.Error(err))
			if ki.discardOnError {
				continue
			}
			data = map[string]any{
				"message": string(msg.Value),
				"_error":  err.Error(),
			}
		}
		if ki.decorateEvents {
			kafkaMeta := map[string]any{
				"topic":     msg.Topic,
				"partition": msg.Partition,
				"offset":    msg.Offset,
				// "timestamp": msg.Timestamp,
			}
			data["@metadata"] = map[string]any{"kafka": kafkaMeta}
		}
		return &model.Event{
			Data: data,
			Time: time.Now(),
		}
	}
	return nil
}

func (ki *KafkaInput) Close() {
	ki.stop = true
	for _, c := range ki.groupConsumers {
		if err := (*c).Close(); err != nil {
			logger.Error("kafka input: groupConsumers closing failed", zap.Error(err))
		}
	}
	logger.Info("kafka input: groupConsumers closed ok")
}

// Setup 在消费者组开始消费前调用，用于初始化
func (k *KafkaInput) Setup(session sarama.ConsumerGroupSession) error {
	logger.Info("kafka input: consumer group setup")
	fmt.Println("session:", session.MemberID(), session.Claims())
	return nil
}

// Cleanup 在消费者组停止消费前调用，用于资源清理
func (k *KafkaInput) Cleanup(sarama.ConsumerGroupSession) error {
	logger.Info("kafka input: consumer group cleanup")
	return nil
}

// ConsumeClaim 处理分配给该消费者的分区消息
func (k *KafkaInput) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	logger.Info("kafka input: consumer start")
	// 从claim的Messages()通道中读取消息
	for msg := range claim.Messages() {
		k.messages <- msg
		// logger.Debug(
		// 	fmt.Sprintf("kafka input: 主题: %s, 分区: %d, 偏移量: %d, 时间戳: %v, 消息内容: %s",
		// 		msg.Topic,
		// 		msg.Partition,
		// 		msg.Offset,
		// 		msg.Timestamp,
		// 		string(msg.Value),
		// 	),
		// )
		// if len(msg.Value) == 0 {
		// 	logger.Error("kafka input: msg value is nil", zap.String("topic", msg.Topic))
		// 	continue
		// }

		// data, err := k.decoder.Decode(msg.Value)
		// if err != nil || data == nil {
		// 	logger.Error("kafka input: decoded msg failed: "+string(msg.Value), zap.Error(err))
		// 	if k.discardOnError {
		// 		continue
		// 	}
		// 	data = map[string]any{
		// 		"message": string(msg.Value),
		// 		"_error":  err.Error(),
		// 	}
		// }

		// if k.decorateEvents {
		// 	data["_kafka"] = map[string]any{
		// 		"topic":     msg.Topic,
		// 		"partition": msg.Partition,
		// 		"offset":    msg.Offset,
		// 		"timestamp": msg.Timestamp,
		// 	}
		// 	// kafkaMeta := make(map[string]any)
		// 	// kafkaMeta["topic"] = message.TopicName
		// 	// kafkaMeta["partition"] = message.PartitionID
		// 	// kafkaMeta["offset"] = message.Message.Offset
		// 	// event["@metadata"] = map[string]any{"kafka": kafkaMeta}
		// }

		// k.ch <- &model.Event{
		// 	Data: data,
		// 	Time: time.Now(),
		// }

		// 手动提交偏移量，标记消息已处理
		// 注意：在生产环境中，应确保消息真正处理完成后再提交偏移量
		// session.MarkMessage(msg, "")
	}
	return nil
}
