package kafka

import (
	"context"
	"github.com/go-slark/slark/logger"
	"github.com/segmentio/kafka-go"
	"github.com/zeromicro/go-zero/core/executors"
	"time"
)

type Producer struct {
	kw       *kafka.Writer
	topic    string
	executor *executors.ChunkExecutor
}

type option struct {
	allowAutoTopicCreation bool
	interval               time.Duration
	chunkSize              int
}

type Option func(*option)

func ChunkSize(size int) Option {
	return func(o *option) {
		o.chunkSize = size
	}
}

func Interval(interval time.Duration) Option {
	return func(o *option) {
		o.interval = interval
	}
}

func AutoTopicCreation(auto bool) Option {
	return func(o *option) {
		o.allowAutoTopicCreation = auto
	}
}

func NewProducer(addr []string, topic string, opts ...Option) *Producer {
	kw := &kafka.Writer{
		Addr:        kafka.TCP(addr...),
		Topic:       topic,
		Balancer:    &kafka.LeastBytes{},
		Compression: kafka.Snappy,
	}
	producer := &Producer{
		kw:    kw,
		topic: topic,
	}
	var o option
	for _, opt := range opts {
		opt(&o)
	}
	kw.AllowAutoTopicCreation = o.allowAutoTopicCreation
	co := make([]executors.ChunkOption, 0)
	if o.chunkSize > 0 {
		co = append(co, executors.WithChunkBytes(o.chunkSize))
	}
	if o.interval > 0 {
		co = append(co, executors.WithFlushInterval(o.interval))
	}
	f := func(tasks []any) {
		messages := make([]kafka.Message, 0, len(tasks))
		var (
			msg kafka.Message
			ok  bool
		)
		for _, task := range tasks {
			msg, ok = task.(kafka.Message)
			if ok {
				messages = append(messages, msg)
			}
		}
		err := producer.kw.WriteMessages(context.TODO(), messages...)
		if err != nil {
			logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err}, "batch produce kafka msg error")
		}
	}
	producer.executor = executors.NewChunkExecutor(f, co...)
	return producer
}

func (p *Producer) Produce(ctx context.Context, k, v []byte) error {
	msg := kafka.Message{
		Key:   k,
		Value: v,
	}
	if p.executor != nil {
		return p.executor.Add(msg, len(v))
	}
	return p.kw.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	if p.executor != nil {
		p.executor.Flush()
	}
	return p.kw.Close()
}
