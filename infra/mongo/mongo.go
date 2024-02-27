package mongo

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/logger"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Config struct {
	Url         string `json:"url"`
	DateBase    string `json:"data_base"`
	Timeout     int    `json:"timeout"`
	MaxPoolSize uint64 `json:"max_pool_size"`
	MinPoolSize uint64 `json:"min_pool_size"`
	Monitor     bool   `json:"monitor"`
	logger.Logger
}

func client(c *Config) (*mongo.Client, error) {
	opts := options.Client()
	if c.MaxPoolSize != 0 {
		opts.SetMaxPoolSize(c.MaxPoolSize)
	}
	if c.MinPoolSize != 0 {
		opts.SetMinPoolSize(c.MinPoolSize)
	}
	if c.Monitor {
		opts.SetMonitor(&event.CommandMonitor{
			Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
				c.Log(ctx, logger.InfoLevel, map[string]interface{}{}, fmt.Sprintf("%v", startedEvent.Command))
			},
			Succeeded: func(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
				c.Log(ctx, logger.InfoLevel, map[string]interface{}{}, fmt.Sprintf("%v", succeededEvent.Reply))
			},
			Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
				c.Log(ctx, logger.InfoLevel, map[string]interface{}{}, fmt.Sprintf("%v", failedEvent.Failure))
			},
		})
	}
	opts.ApplyURI(c.Url)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout)*time.Second)
	defer cancel()
	cli, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	err = cli.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

func New(cfg *Config) (*Client, error) {
	cli, err := client(cfg)
	if err != nil {
		return nil, err
	}
	c := &Client{
		client: cli,
		db:     cli.Database(cfg.DateBase),
	}
	return c, nil
}

func (c *Client) Coll(coll string) *mongo.Collection {
	return c.db.Collection(coll)
}

func (c *Client) Database() *mongo.Database {
	return c.db
}

func (c *Client) Close() error {
	return c.client.Disconnect(context.TODO())
}
