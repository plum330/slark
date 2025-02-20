package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Subscription struct {
	gorm.Model
	UserId   string `gorm:"type:varchar(20);not null"`
	PubKey   string `gorm:"type:varchar(128)"`
	PriKey   string `gorm:"type:varchar(64)"`
	Endpoint string `gorm:"type:varchar(1024)"`
	Auth     string `gorm:"type:varchar(32)"`
	P256dh   string `gorm:"type:varchar(128)"`
}

type State struct {
	db    *gorm.DB
	cache *Cache
}

func NewState(db *gorm.DB, redis *redis.Client) *State {
	return &State{
		db:    db,
		cache: New(redis, Error(gorm.ErrRecordNotFound)),
	}
}

func (w *State) Create(ctx context.Context, s *Subscription) error {
	return w.db.Create(s).Error
}

func (w *State) Update(ctx context.Context, s *Subscription) error {
	ss, err := w.Find(ctx, s.UserId)
	if err != nil {
		return err
	}
	pKey := fmt.Sprintf("%d|subscription_notify", ss.ID)
	uKey := fmt.Sprintf("%s|subscription_notify", s.UserId)
	return w.cache.Exec(ctx, s, func(v any) error {
		x, ok := v.(*Subscription)
		if !ok {
			return errors.New("invalid type assert")
		}

		uid := x.UserId
		x.UserId = ""
		return w.db.Model(Subscription{}).Where("user_id=?", uid).Updates(x).Error
	}, pKey, uKey)
}

func (w *State) Find(ctx context.Context, uid string) (*Subscription, error) {
	s := &Subscription{}
	// 唯一key
	uKey := fmt.Sprintf("%s|subscription_notify", uid)
	err := w.cache.FetchIndex(ctx, s, uKey, func(v any) string {
		// 主键key
		return fmt.Sprintf("%v|subscription_notify", v)
	}, func(v any) (any, error) {
		// 查询主键
		ss := &Subscription{}
		err := w.db.Model(Subscription{}).Where("user_id=?", uid).First(ss).Error
		if err != nil {
			return nil, err
		}
		*v.(*Subscription) = *ss
		return ss.ID, nil
	}, func(pk, v any) error {
		// 通过主键查询
		return w.db.Model(Subscription{}).Where("id=?", pk).First(v).Error
	})
	if err != nil {
		return nil, err
	}
	return s, nil
}
