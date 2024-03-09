package limit

import (
	"context"
	"github.com/go-slark/slark/errors"
	"github.com/go-slark/slark/logger"
)

type MaxConn struct {
	pool *Pool
}

func NewMaxConn() *MaxConn {
	return &MaxConn{pool: NewPool(10000)}
}

func (c *MaxConn) Pass() error {
	allow := c.pool.Use()
	if !allow {
		return errors.ServerUnavailable("conn overload", "CONN_OVERLOAD")
	}
	err := c.pool.Back()
	if err != nil {
		logger.Log(context.TODO(), logger.ErrorLevel, map[string]interface{}{"error": err})
	}
	return nil
}
