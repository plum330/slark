package grpc

import (
	"github.com/go-slark/slark/errors"
	"google.golang.org/grpc"
	"os"
)

// GRPC Server

type RegisterObj struct {
	Obj      interface{}
	Register func(s *grpc.Server, obj interface{})
}

type RegisterObjSet struct {
	Sets []RegisterObj
}

func (r *RegisterObjSet) NewGRPCServer(opts ...ServerOption) *Server {
	srv := NewServer(opts...)
	for _, set := range r.Sets {
		set.Register(srv.Server, set.Obj)
	}
	return srv
}

// GRPC Client

type GRPCClient struct {
	clients map[string]*grpc.ClientConn
}

type ClientObj struct {
	Name string
	Addr string
}

func NewGRPCClient(objs []*ClientObj, opt []grpc.DialOption, opts ...ClientOption) *GRPCClient {
	clients := make(map[string]*grpc.ClientConn, len(objs))
	for _, obj := range objs {
		client := NewClient(append(append(append([]ClientOption{}, WithAddr(obj.Addr)), ClientOptions(opt)), opts...)...)
		if client.err != nil {
			os.Exit(errors.ClientClosed)
		}
		clients[obj.Name] = client.ClientConn
	}
	return &GRPCClient{clients: clients}
}

func (c *GRPCClient) GetGRPCClient(name string) *grpc.ClientConn {
	return c.clients[name]
}

func (c *GRPCClient) Stop() {
	for _, client := range c.clients {
		_ = client.Close()
	}
}
