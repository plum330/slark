package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/pkg/routine"
	tracing "github.com/go-slark/slark/pkg/trace"
	"github.com/zhenjl/cityhash"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"time"
)

const (
	msgTopic = "msg_topic"
	msgKey   = "msg_key"
	msgValue = "msg_value"
)

type KafkaProducer struct {
	sarama.SyncProducer
	sarama.AsyncProducer
	logger.Logger
	*tracing.Tracer
}

type ProducerConf struct {
	Brokers       []string `mapstructure:"brokers"`
	Retry         int      `mapstructure:"retry"`
	Ack           int16    `mapstructure:"ack"`
	ReturnSuccess bool     `mapstructure:"return_success"`
	ReturnErrors  bool     `mapstructure:"return_errors"`
	Version       string   `mapstructure:"version"`
	TraceEnable   bool     `mapstructure:"trace_enable"`
}

type ConsumerGroupConf struct {
	Brokers      []string      `mapstructure:"brokers"`
	GroupID      string        `mapstructure:"group_id"`
	Topics       []string      `mapstructure:"topics"`
	Initial      int64         `mapstructure:"initial"`
	ReturnErrors bool          `mapstructure:"return_errors"`
	AutoCommit   bool          `mapstructure:"auto_commit"`
	Interval     time.Duration `mapstructure:"interval"`
	WorkerNum    uint          `mapstructure:"worker_num"`
	Version      string        `mapstructure:"version"`
	Worker       int           `mapstructure:"worker"`
	TraceEnable  bool          `mapstructure:"trace_enable"`
}

type KafkaConf struct {
	Producer      *ProducerConf      `mapstructure:"producer"`
	ConsumerGroup *ConsumerGroupConf `mapstructure:"consumer_group"`
}

func (kp *KafkaProducer) Close() {
	_ = kp.SyncProducer.Close()
	kp.AsyncClose()
}

func (kp *KafkaProducer) SyncSend(ctx context.Context, topic, key string, msg []byte) error {
	pm := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Key:   sarama.StringEncoder(key),
	}
	if kp.Tracer != nil {
		opt := []trace.SpanStartOption{
			trace.WithSpanKind(kp.Kind()),
			trace.WithAttributes(attribute.String(msgTopic, topic), attribute.String(msgKey, key), attribute.String(msgValue, string(msg))),
		}
		_, span := kp.Start(ctx, "kafka sync send", &producerMsgCarrier{pm}, opt...)
		defer span.End()
	}
	_, _, err := kp.SyncProducer.SendMessage(pm)
	return err
}

func (kp *KafkaProducer) AsyncSend(ctx context.Context, topic, key string, msg []byte) error {
	pm := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Key:   sarama.StringEncoder(key),
	}
	if kp.Tracer != nil {
		opt := []trace.SpanStartOption{
			trace.WithSpanKind(kp.Kind()),
			trace.WithAttributes(attribute.String(msgTopic, topic), attribute.String(msgKey, key), attribute.String(msgValue, string(msg))),
		}
		x, span := kp.Start(ctx, "kafka async send", &producerMsgCarrier{pm}, opt...)
		pm.Metadata = x
		defer span.End()
	}
	kp.AsyncProducer.Input() <- pm
	return nil
}

func (kp *KafkaProducer) monitor() {
	var (
		msg *sarama.ProducerMessage
		e   *sarama.ProducerError
	)
	go func(ap sarama.AsyncProducer) {
		for msg = range ap.Successes() {
			if msg != nil {
				value, _ := msg.Value.Encode()
				ctx, _ := msg.Metadata.(context.Context)
				kp.Log(ctx, logger.DebugLevel, map[string]interface{}{"topic": msg.Topic, "key": msg.Key, "value": string(value)}, "kafka async produce msg success")
			}
		}
	}(kp.AsyncProducer)

	go func(ap sarama.AsyncProducer) {
		for e = range ap.Errors() {
			if e == nil {
				continue
			}

			if e.Msg != nil {
				value, _ := e.Msg.Value.Encode()
				ctx, _ := e.Msg.Metadata.(context.Context)
				kp.Log(ctx, logger.ErrorLevel, map[string]interface{}{"error": e.Err, "topic": e.Msg.Topic, "key": e.Msg.Key, "value": string(value)}, "kafka async produce msg fail")
			} else {
				kp.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": e.Err}, "kafka async produce msg fail")
			}
		}
	}(kp.AsyncProducer)
}

func NewKafkaProducer(conf *ProducerConf, opts ...tracing.Option) (*KafkaProducer, error) {
	sp, err := newSyncProducer(conf)
	if err != nil {
		return nil, err
	}
	ap, err := newAsyncProducer(conf)
	if err != nil {
		return nil, err
	}
	kp := &KafkaProducer{
		SyncProducer:  sp,
		AsyncProducer: ap,
		Logger:        logger.GetLogger(),
	}
	if conf.TraceEnable {
		kp.Tracer = tracing.NewTracer(trace.SpanKindProducer, opts...)
	}
	kp.monitor()
	return kp, nil
}

func newSyncProducer(conf *ProducerConf) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.RequiredAcks(conf.Ack) // WaitForAll
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Retry.Max = conf.Retry
	config.Producer.Return.Successes = true // true
	//config.Producer.Return.Errors = true     // default true
	version, err := sarama.ParseKafkaVersion(conf.Version)
	if err != nil {
		return nil, err
	}
	config.Version = version
	if err = config.Validate(); err != nil {
		return nil, err
	}

	return sarama.NewSyncProducer(conf.Brokers, config)
}

func newAsyncProducer(conf *ProducerConf) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.RequiredAcks(conf.Ack) // WaitForAll
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Retry.Max = conf.Retry
	config.Producer.Return.Successes = conf.ReturnSuccess // true / false
	config.Producer.Return.Errors = conf.ReturnErrors     // true / false
	version, err := sarama.ParseKafkaVersion(conf.Version)
	if err != nil {
		return nil, err
	}
	config.Version = version
	if err = config.Validate(); err != nil {
		return nil, err
	}

	return sarama.NewAsyncProducer(conf.Brokers, config)
}

type Consume interface {
	Handler(context.Context, *sarama.ConsumerMessage) error
}

type KafkaConsumerGroup struct {
	sarama.ConsumerGroup
	sarama.ConsumerGroupHandler
	logger.Logger
	*tracing.Tracer
	topics   []string
	ctx      context.Context
	cf       context.CancelFunc
	handlers map[string]Consume
	worker   int
	chs      []chan *sarama.ConsumerMessage
}

func NewKafkaConsumer(conf *ConsumerGroupConf, opts ...tracing.Option) (*KafkaConsumerGroup, error) {
	cg, err := newConsumerGroup(conf)
	if err != nil {
		return nil, err
	}
	k := &KafkaConsumerGroup{
		ConsumerGroup: cg,
		topics:        conf.Topics,
		Logger:        logger.GetLogger(),
		handlers:      make(map[string]Consume),
		worker:        conf.Worker,
		chs:           make([]chan *sarama.ConsumerMessage, conf.Worker),
	}
	k.ConsumerGroupHandler = k
	for i := 0; i < k.worker; i++ {
		ch := make(chan *sarama.ConsumerMessage, 1024)
		k.chs[i] = ch
		routine.GoSafe(context.TODO(), func() {
			k.consume(ch)
		})
	}
	if conf.TraceEnable {
		k.Tracer = tracing.NewTracer(trace.SpanKindConsumer, opts...)
	}
	k.ctx, k.cf = context.WithCancel(context.TODO())
	return k, nil
}

func newConsumerGroup(conf *ConsumerGroupConf) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = conf.Initial              // -2:sarama.OffsetOldest -1:sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = conf.AutoCommit // false
	config.Consumer.Offsets.AutoCommit.Interval = conf.Interval * time.Millisecond
	config.Consumer.Return.Errors = conf.ReturnErrors // true
	version, err := sarama.ParseKafkaVersion(conf.Version)
	if err != nil {
		return nil, err
	}
	config.Version = version
	if err = config.Validate(); err != nil {
		return nil, err
	}

	return sarama.NewConsumerGroup(conf.Brokers, conf.GroupID, config)
}

func (k *KafkaConsumerGroup) Register(topic string, handler Consume) {
	k.handlers[topic] = handler
}

func (*KafkaConsumerGroup) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*KafkaConsumerGroup) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (k *KafkaConsumerGroup) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		index := cityhash.CityHash32(msg.Key, uint32(len(msg.Key))) % uint32(k.worker)
		k.chs[index] <- msg
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (k *KafkaConsumerGroup) Consume() {
	for {
		err := k.ConsumerGroup.Consume(k.ctx, k.topics, k.ConsumerGroupHandler)
		if err != nil {
			k.Log(k.ctx, logger.WarnLevel, map[string]interface{}{"error": err}, "consumer group consume fail")
		}
		if k.ctx.Err() != nil {
			k.Log(k.ctx, logger.ErrorLevel, map[string]interface{}{"error": k.ctx.Err()}, "consumer group exit")
			return
		}
		time.Sleep(time.Second)
	}
}

func (k *KafkaConsumerGroup) consume(ch <-chan *sarama.ConsumerMessage) {
	for {
		msg := <-ch
		handler, ok := k.handlers[msg.Topic]
		if !ok {
			k.Log(context.TODO(), logger.WarnLevel, map[string]interface{}{"topic": msg.Topic}, "topic unregister")
			continue
		}

		ctx := context.Background()
		var span trace.Span
		if k.Tracer != nil {
			opt := []trace.SpanStartOption{
				trace.WithSpanKind(k.Kind()),
				trace.WithAttributes(attribute.String(msgTopic, msg.Topic), attribute.String(msgKey, string(msg.Key)), attribute.String(msgValue, string(msg.Value))),
			}
			ctx, span = k.Tracer.Start(context.TODO(), "kafka group consume", &consumerMsgCarrier{msg}, opt...)
			span.End()
		}
		err := handler.Handler(ctx, msg)
		if err != nil {
			k.Log(ctx, logger.ErrorLevel, map[string]interface{}{"error": err}, "handle consume msg error")
		}
	}
}

func (k *KafkaConsumerGroup) Start() error {
	k.Consume()
	return nil
}

func (k *KafkaConsumerGroup) Stop(_ context.Context) error {
	k.cf()
	return k.Close()
}

var (
	kafkaProducer      *KafkaProducer
	kafkaConsumerGroup *KafkaConsumerGroup
)

func GetKafkaProducer() *KafkaProducer {
	return kafkaProducer
}

func GetKafkaConsumerGroup() *KafkaConsumerGroup {
	return kafkaConsumerGroup
}

type producerMsgCarrier struct {
	msg *sarama.ProducerMessage
}

func (c *producerMsgCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *producerMsgCarrier) Set(key, val string) {
	for i := 0; i < len(c.msg.Headers); i++ {
		if string(c.msg.Headers[i].Key) == key {
			c.msg.Headers = append(c.msg.Headers[:i], c.msg.Headers[i+1:]...)
			i--
		}
	}
	c.msg.Headers = append(c.msg.Headers, sarama.RecordHeader{
		Key:   []byte(key),
		Value: []byte(val),
	})
}

func (c *producerMsgCarrier) Keys() []string {
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = string(h.Key)
	}
	return out
}

type consumerMsgCarrier struct {
	msg *sarama.ConsumerMessage
}

func (c *consumerMsgCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if h != nil && string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c *consumerMsgCarrier) Set(key, val string) {
	for i := 0; i < len(c.msg.Headers); i++ {
		if c.msg.Headers[i] != nil && string(c.msg.Headers[i].Key) == key {
			c.msg.Headers = append(c.msg.Headers[:i], c.msg.Headers[i+1:]...)
			i--
		}
	}
	c.msg.Headers = append(c.msg.Headers, &sarama.RecordHeader{
		Key:   []byte(key),
		Value: []byte(val),
	})
}

func (c *consumerMsgCarrier) Keys() []string {
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = string(h.Key)
	}
	return out
}
