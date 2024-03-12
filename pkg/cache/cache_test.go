package cache

import (
	"context"
	"fmt"
	"github.com/go-slark/slark/infra/mysql"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
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
	}), Error(redis.Nil), Expiry(5*time.Minute))
	v := &Value{}
	_, err := c.Fetch(context.TODO(), "fetch", v, func(value any) error {
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
	}), Error(redis.Nil), Expiry(5*time.Minute))
	var v string
	_, err := c.Fetch(context.TODO(), "fetch_str", &v, func(value any) error {
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
	}), Error(redis.Nil), Expiry(3*time.Minute))
	var str string
	err := c.FetchIndex(context.TODO(), "fetch_index_str", func(k any) string {
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
	}), Error(redis.Nil), Expiry(3*time.Minute))
	value := &Value{}
	err := c.FetchIndex(context.TODO(), "fetch_index", func(k any) string {
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

type AdminModel struct {
	gorm.Model
	State    int
	Role     int
	Name     string
	Account  string
	Password string
}

func TestFromDB(t *testing.T) {
	client, err := mysql.New(&mysql.Config{
		Address:     "",
		MaxIdleConn: 5,
		MaxOpenConn: 20,
		MaxLifeTime: 300,
		LogMode:     4,
	})
	if err != nil {
		t.Fatalf("new db error:%+v", err)
	}
	usr := &AdminModel{}
	c := New(redis.NewClient(&redis.Options{
		Addr:     "",
		Password: "",
		DB:       10,
	}), Expiry(5*time.Minute))
	_, err = c.Fetch(context.TODO(), "primary-key:1", usr, func(v any) error {
		return client.Model(AdminModel{}).Where("id = 1").First(v).Error
	})
	if err != nil {
		t.Errorf("fetch error:%+v", err)
	}
	fmt.Printf("primary-key:1:%+v\n", usr)
	// db不存在
	_, err = c.Fetch(context.TODO(), "primary-key:0", usr, func(v any) error {
		return client.Model(AdminModel{}).Where("id = ?", 0).First(v).Error
	})
	if err != nil {
		t.Errorf("fetch error:%+v", err)
	}
	fmt.Printf("primary-key:0:%+v\n", usr)
	err = c.FetchIndex(context.TODO(), "unique-key:account", func(v any) string {
		return fmt.Sprintf("primary-key:%v", v)
	}, usr, func(v any) error {
		err = client.Model(AdminModel{}).Where("account = ?", "YY01").First(usr).Error
		if err == nil {
			// 设置primary key
			*v.(*any) = usr.ID
		}
		return err
	}, func(v any) error {
		return client.Model(AdminModel{}).Where("account = ?", "YY01").First(v).Error
	})
	if err != nil {
		t.Errorf("fetch index error:%+v", err)
	}
	fmt.Printf("unique-key:account:%+v\n", usr)
}
