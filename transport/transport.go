package transport

import (
	"context"

	_ "github.com/go-slark/slark/encoding/form"
	_ "github.com/go-slark/slark/encoding/json"
	_ "github.com/go-slark/slark/encoding/msgpack"
	_ "github.com/go-slark/slark/encoding/proto"
)

type Server interface {
	Start() error
	Stop(ctx context.Context) error
}
