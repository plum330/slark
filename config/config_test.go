package config

import (
	"github.com/go-slark/slark/config/source/env"
	"os"
	"sync"
	"testing"
)

type Redis struct {
	Addr    string `json:"addr"`
	Timeout int    `json:"timeout"`
}

type Mongo struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type KafkaProducer struct {
	Addr string `json:"addr"`
	Port int    `json:"port"`
}

type Kafka struct {
	Producer KafkaProducer `json:"producer"`
}

type AllConfig struct {
	Redis Redis `json:"redis"`
	Mongo Mongo `json:"mongo"`
	Kafka Kafka `json:"kafka"`
}

func TestFileConfig(t *testing.T) {
	c := New()
	c.callback = append(c.callback, func(m *sync.Map) {
		v, _ := c.cached.Load("redis.addr")
		t.Logf("config:%+v", v)
	})
	err := c.Load()
	if err != nil {
		t.Fatalf("load config error:%+v", err)
	}

	r := &Redis{}
	err = c.Unmarshal(r, "redis")
	if err != nil {
		t.Fatalf("unmarshal error:%+v", err)
	}
	t.Logf("redis config:%+v", r)

	ac := &AllConfig{}
	err = c.Unmarshal(ac)
	if err != nil {
		t.Fatalf("++ unmarshal error:%+v", err)
	}
	t.Logf("all config:%+v", ac)

	<-make(chan struct{})
}

type RedisConfig struct {
	RedisAddr string `json:"redis_addr"`
}

func TestEnvConfig(t *testing.T) {
	os.Setenv("slark_redis_addr", "redis://127.0.0.1:8080")
	c := New(WithSource(env.New()))
	c.callback = append(c.callback, func(m *sync.Map) {
		v, _ := c.cached.Load("redis_addr")
		t.Logf("config:%+v", v)
	})
	err := c.Load()
	if err != nil {
		t.Fatalf("load config error:%+v", err)
	}

	r := &RedisConfig{}
	err = c.Unmarshal(r)
	if err != nil {
		t.Fatalf("unmarshal redis config error:%+v", err)
	}
	t.Logf("redis_addr:%s", r.RedisAddr)
	select {}
}
