package mysql

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Option func(db *gorm.DB)

func NotINT64Zero(param string, data int64) Option {
	return func(db *gorm.DB) {
		if data == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s = ?", param), data)
	}
}

func NotUINT64Zero(param string, data uint64) Option {
	return func(db *gorm.DB) {
		if data == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s = ?", param), data)
	}
}

func NotUint32Zero(param string, data uint32) Option {
	return func(db *gorm.DB) {
		if data == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s = ?", param), data)
	}
}

func NotInt8Zero(param string, data int8) Option {
	return func(db *gorm.DB) {
		if data == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s = ?", param), data)
	}
}

func UINT64(param string, data uint64) Option {
	return func(db *gorm.DB) {
		db.Where(fmt.Sprintf("%s = ?", param), data)
	}
}

func INT64(param string, data int64) Option {
	return func(db *gorm.DB) {
		db.Where(fmt.Sprintf("%s = ?", param), data)
	}
}

func STRING(param, data string) Option {
	return func(db *gorm.DB) {
		if len(data) == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s = ?", param), data)
	}
}

func StringIn(param string, data []string) Option {
	return func(db *gorm.DB) {
		if len(data) == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s in (?)", param), data)
	}
}

func Uint64In(param string, data []uint64) Option {
	return func(db *gorm.DB) {
		if len(data) == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s in (?)", param), data)
	}
}

func Int64In(param string, data []int64) Option {
	return func(db *gorm.DB) {
		if len(data) == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s in (?)", param), data)
	}
}

func Uint32In(param string, data []uint32) Option {
	return func(db *gorm.DB) {
		if len(data) == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s in (?)", param), data)
	}
}

func Time(param string, sign string, t time.Time) Option {
	return func(db *gorm.DB) {
		if t.Unix() == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s %s ?", param, sign), t)
	}
}

func LIKE(param string, data string) Option {
	return func(db *gorm.DB) {
		if len(data) == 0 {
			return
		}
		db.Where(fmt.Sprintf("%s like ?", param), fmt.Sprintf("%%%s%%", data))
	}
}

func LIMIT(limit int) Option {
	return func(db *gorm.DB) {
		if limit <= 0 {
			return
		}
		db.Limit(limit)
	}
}

func OFFSET(offset int) Option {
	return func(db *gorm.DB) {
		if offset <= 0 {
			return
		}
		db.Offset(offset)
	}
}

func COUNT(total *int64) Option {
	return func(db *gorm.DB) {
		if total == nil {
			return
		}
		db.Count(total)
	}
}

func ORDER(order string) Option {
	return func(db *gorm.DB) {
		db.Order(order)
	}
}

func GROUP(name string) Option {
	return func(db *gorm.DB) {
		db.Group(name)
	}
}

func WHERE(query string, arg interface{}) Option {
	return func(db *gorm.DB) {
		db.Where(query, arg)
	}
}

// -------------------

type QueryOption struct {
	Skip  int
	Limit int
	Order []string
}

type QueryOptFunc func(*QueryOption)

func ApplyQueryOpts(db *gorm.DB, opts ...QueryOptFunc) *gorm.DB {
	queryOption := &QueryOption{}
	for _, opt := range opts {
		opt(queryOption)
	}

	for _, order := range queryOption.Order {
		db = db.Order(order)
	}

	if queryOption.Limit != 0 {
		db = db.Limit(queryOption.Limit)
	}

	if queryOption.Skip != 0 {
		db = db.Offset(queryOption.Skip)
	}
	return db
}

func Skip(skip int) QueryOptFunc {
	return func(option *QueryOption) {
		option.Skip = skip
	}
}

func Limit(limit int) QueryOptFunc {
	return func(option *QueryOption) {
		option.Limit = limit
	}
}

func Order(order ...string) QueryOptFunc {
	return func(option *QueryOption) {
		option.Order = append(option.Order, order...)
	}
}
