package mysql

import (
	xlogger "github.com/go-slark/slark/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/opentelemetry/tracing"
	"time"
)

type Config struct {
	Address       string `json:"address"`
	MaxIdleConn   int    `json:"max_idle_conn"`
	MaxOpenConn   int    `json:"max_open_conn"`
	MaxLifeTime   int    `json:"max_life_time"`
	MaxIdleTime   int    `json:"max_idle_time"`
	LogMode       int    `json:"log_mode"` //默认warn
	CustomizedLog bool   `json:"customized_log"`
	xlogger.Logger
}

type Client struct {
	*gorm.DB
}

func New(c *Config) (*Client, error) {
	var l logger.Interface
	if c.CustomizedLog {
		l = newCustomizedLogger(WithLogLevel(logger.LogLevel(c.LogMode)), WithLogger(c.Logger))
	} else {
		l = logger.Default.LogMode(logger.LogLevel(c.LogMode))
	}
	cfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		Logger:         l,
	}
	db, err := gorm.Open(mysql.Open(c.Address), cfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConn)
	sqlDB.SetMaxOpenConns(c.MaxOpenConn)
	if c.MaxLifeTime != 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(c.MaxLifeTime) * time.Second)
	}
	if c.MaxIdleTime != 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(c.MaxIdleTime) * time.Second)
	}

	if err = sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics()))
	if err != nil {
		return nil, err
	}
	return &Client{DB: db}, nil
}

func (c *Client) Database() *gorm.DB {
	return c.DB
}

func (c *Client) Close() {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}
