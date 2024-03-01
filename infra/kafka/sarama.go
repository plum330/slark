package kafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/go-slark/slark/logger"
	"github.com/go-slark/slark/pkg"
	"github.com/go-slark/slark/pkg/routine"
	tracing "github.com/go-slark/slark/pkg/trace"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
	"time"
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
	traceID, ok := ctx.Value(utils.RayID).(string)
	var spanID string
	if !ok {
		traceID = tracing.ExtractTraceID(ctx)
		spanID = tracing.ExtractSpanID(ctx)
	}

	pm := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Key:   sarama.StringEncoder(key),
		Headers: []sarama.RecordHeader{
			{
				Key:   sarama.ByteEncoder(utils.RayID),
				Value: sarama.ByteEncoder(traceID),
			},
			{
				Key:   sarama.ByteEncoder(utils.SpanID),
				Value: sarama.ByteEncoder(spanID),
			},
		},
	}
	if kp.Tracer != nil {
		opt := []trace.SpanStartOption{
			trace.WithSpanKind(kp.Kind()),
			trace.WithAttributes(attribute.String("mq_topic", topic), attribute.String("mq_key", key), attribute.String("mq_msg", string(msg))),
		}
		carrier := &tracing.Carrier{MD: &metadata.MD{}}
		_, span := kp.Start(ctx, "kafka sync send", carrier, opt...)
		for _, k := range carrier.Keys() {
			pm.Headers = append(pm.Headers, sarama.RecordHeader{
				Key:   sarama.ByteEncoder(k),
				Value: sarama.ByteEncoder(carrier.Get(k)),
			})
		}
		defer span.End()
	}
	_, _, err := kp.SyncProducer.SendMessage(pm)
	return err
}

func (kp *KafkaProducer) AsyncSend(ctx context.Context, topic, key string, msg []byte) error {
	traceID, ok := ctx.Value(utils.RayID).(string)
	var spanID string
	if !ok {
		traceID = tracing.ExtractTraceID(ctx)
		spanID = tracing.ExtractSpanID(ctx)
	}

	pm := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg),
		Key:   sarama.StringEncoder(key),
		Headers: []sarama.RecordHeader{
			{
				Key:   sarama.ByteEncoder(utils.RayID),
				Value: sarama.ByteEncoder(traceID),
			},
			{
				Key:   sarama.ByteEncoder(utils.SpanID),
				Value: sarama.ByteEncoder(spanID),
			},
		},
	}
	if kp.Tracer != nil {
		opt := []trace.SpanStartOption{
			trace.WithSpanKind(kp.Kind()),
			trace.WithAttributes(attribute.String("mq_topic", topic), attribute.String("mq_key", key), attribute.String("mq_msg", string(msg))),
		}
		carrier := &tracing.Carrier{MD: &metadata.MD{}}
		_, span := kp.Start(ctx, "kafka async send", carrier, opt...)
		for _, k := range carrier.Keys() {
			pm.Headers = append(pm.Headers, sarama.RecordHeader{
				Key:   sarama.ByteEncoder(k),
				Value: sarama.ByteEncoder(carrier.Get(k)),
			})
		}
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
				kp.Log(context.TODO(), logger.DebugLevel, map[string]interface{}{"topic": msg.Topic, "key": msg.Key, "value": msg.Value}, "kafka async produce msg succ")
			}
		}
	}(kp.AsyncProducer)

	go func(ap sarama.AsyncProducer) {
		for e = range ap.Errors() {
			if e == nil {
				continue
			}

			if e.Msg != nil {
				kp.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": e.Err, "topic": e.Msg.Topic, "key": e.Msg.Key, "value": e.Msg.Value}, "kafka async produce msg fail")
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
	Topics []string
	context.Context
	context.CancelFunc
	logger.Logger
	*tracing.Tracer
	handlers map[string]Consume
	worker   chan struct{}
}

func NewKafkaConsumer(conf *ConsumerGroupConf, opts ...tracing.Option) (*KafkaConsumerGroup, error) {
	cg, err := newConsumerGroup(conf)
	if err != nil {
		return nil, err
	}
	k := &KafkaConsumerGroup{
		ConsumerGroup: cg,
		Topics:        conf.Topics,
		Logger:        logger.GetLogger(),
		handlers:      make(map[string]Consume),
		worker:        make(chan struct{}, conf.Worker),
	}
	if conf.TraceEnable {
		k.Tracer = tracing.NewTracer(trace.SpanKindConsumer, opts...)
	}
	k.Context, k.CancelFunc = context.WithCancel(context.TODO())
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

func (kc *KafkaConsumerGroup) Register(topic string, handler Consume) {
	kc.handlers[topic] = handler
}

func (*KafkaConsumerGroup) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*KafkaConsumerGroup) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (kc *KafkaConsumerGroup) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	handler, ok := kc.handlers[claim.Topic()]
	if !ok {
		return nil
	}

	for msg := range claim.Messages() {
		kc.worker <- struct{}{}
		m := msg
		ctx := context.Background()
		var span trace.Span
		routine.GoSafe(ctx, func() {
			defer func() {
				<-kc.worker
			}()
			if kc.Tracer != nil {
				opt := []trace.SpanStartOption{
					trace.WithSpanKind(kc.Kind()),
					trace.WithAttributes(attribute.String("mq_topic", m.Topic), attribute.String("mq_key", string(m.Key)), attribute.String("mq_msg", string(m.Value))),
				}
				md := make(metadata.MD)
				ctx, span = kc.Tracer.Start(ctx, "kafka group consume", &tracing.Carrier{MD: &md}, opt...)
				defer span.End()
			}
			err := handler.Handler(ctx, m)
			if err != nil {
				if span != nil {
					// TODO
				}
			}
		})
		sess.MarkMessage(m, "")
	}
	return nil
}

func (kc *KafkaConsumerGroup) Consume() {
	for {
		err := kc.ConsumerGroup.Consume(kc.Context, kc.Topics, kc.ConsumerGroupHandler)
		if err != nil {
			kc.Log(kc.Context, logger.WarnLevel, map[string]interface{}{"error": err}, "consumer group consume fail")
		}
		if kc.Context.Err() != nil {
			kc.Log(kc.Context, logger.ErrorLevel, map[string]interface{}{"error": kc.Context.Err()}, "consumer group exit")
			return
		}
		time.Sleep(time.Second)
	}
}

func (kc *KafkaConsumerGroup) Start() error {
	kc.Consume()
	return nil
}

func (kc *KafkaConsumerGroup) Stop(_ context.Context) error {
	kc.CancelFunc()
	return kc.Close()
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
