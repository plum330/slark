package cache

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/pkg/sf"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

type Value struct {
	V string `json:"v"`
}

func TestFetch(t *testing.T) {
	c := New(redis.NewClient(&redis.Options{
		Addr:     "192.168.3.13:2379",
		Password: "CtHHQNbFkXpw33ew",
		DB:       10,
	}), sf.NewGroup(), redis.Nil)
	v := &Value{}
	err := c.Fetch(context.TODO(), "fetch", 5*time.Minute, v, func(value any) error {
		vv, _ := value.(*Value)
		vv.V = "***************"
		return nil
	})
	if err != nil {
		t.Errorf("fetch error:%+v", err)
		return
	}
	t.Logf("fetch result:%s", v.V)
}

func TestFetchStr(t *testing.T) {
	c := New(redis.NewClient(&redis.Options{
		Addr:     "192.168.3.13:2379",
		Password: "CtHHQNbFkXpw33ew",
		DB:       10,
	}), sf.NewGroup(), redis.Nil)
	var v string
	err := c.Fetch(context.TODO(), "fetch_str", 5*time.Minute, &v, func(value any) error {
		*value.(*string) = "+++++++++++++++++"
		return nil
	})
	if err != nil {
		t.Errorf("fetch error:%+v", err)
		return
	}
	t.Logf("fetch result:%s", v)
}

func TestFetchIndexStr(t *testing.T) {
	c := New(redis.NewClient(&redis.Options{
		Addr:     "192.168.3.13:2379",
		Password: "CtHHQNbFkXpw33ew",
		DB:       10,
	}), sf.NewGroup(), redis.Nil)
	var str string
	err := c.FetchIndex(context.TODO(), "fetch_index_str", 3*time.Minute, func(k any) string {
		return fmt.Sprintf("fetch_primary_str:%v", k)
	}, &str, func(v any) error {
		//赋值pk
		*v.(*any) = "====="
		return nil
	}, func(v any) error {
		*v.(*string) = "---------------"
		return nil
	})
	if err != nil {
		t.Errorf("fetch index error:%+v", err)
		return
	}
	t.Logf("fetch index result:%v", str)
}

func TestFetchIndex(t *testing.T) {
	c := New(redis.NewClient(&redis.Options{
		Addr:     "192.168.3.13:2379",
		Password: "CtHHQNbFkXpw33ew",
		DB:       10,
	}), sf.NewGroup(), redis.Nil)
	value := &Value{}
	err := c.FetchIndex(context.TODO(), "fetch_index", 3*time.Minute, func(k any) string {
		return fmt.Sprintf("fetch_primary:%v", k)
	}, value, func(v any) error {
		// 赋值pk
		*v.(*any) = &Value{V: "66666"}
		return nil
	}, func(v any) error {
		vv, _ := v.(*Value)
		vv.V = "777777"
		return nil
	})
	if err != nil {
		t.Errorf("fetch index error:%+v", err)
		return
	}
	t.Logf("fetch index result:%v", value.V)
}
