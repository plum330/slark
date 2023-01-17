package grpc

import (
	"google.golang.org/grpc"
	"os"
)

// GRPC Server

type RegisterObj struct {
	Obj      interface{}
	Register func(s *grpc.Server, obj interface{})
}

func (r *RegisterObj) NewGRPCServer(opts ...ServerOption) *Server {
	srv := NewServer(opts...)
	r.Register(srv.Server, r.Obj)
	return srv
}

// GRPC Client

type GRPCClient struct {
	clients map[string]*Client
}

type ClientObj struct {
	Name string
	Addr string
}

type DialOption func() []grpc.DialOption

func NewGRPCClient(objs []*ClientObj, f DialOption, opts ...ClientOption) *GRPCClient {
	clients := make(map[string]*Client, len(objs))
	for _, obj := range objs {
		client := NewClient(append(append(append([]ClientOption{}, WithAddr(obj.Addr)), ClientOptions(f())), opts...)...)
		if client.err != nil {
			os.Exit(800)
		}
		clients[obj.Name] = client
	}
	return &GRPCClient{clients: clients}
}

func (c *GRPCClient) GetGRPCClient(name string) *Client {
	return c.clients[name]
}

func (c *GRPCClient) Stop() {
	for _, client := range c.clients {
		_ = client.Stop()
	}
}
