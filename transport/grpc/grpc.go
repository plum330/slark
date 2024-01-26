package grpc

import (
	"google.golang.org/grpc"
	"time"
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

type Client struct {
	clients map[string]*grpc.ClientConn
}

type ClientObj struct {
	Name   string
	Addr   string
	Timout time.Duration
}

func NewClient(objs []*ClientObj, opts ...Option) *Client {
	clients := make(map[string]*grpc.ClientConn, len(objs))
	for _, obj := range objs {
		if obj.Timout == 0 {
			obj.Timout = 5 * time.Second
		}
		opts = append(opts, WithAddr(obj.Addr), WithUnaryInterceptor([]grpc.UnaryClientInterceptor{UnaryClientTimeout(obj.Timout)}))
		client, err := Dial(opts...)
		if err != nil {
			panic(err)
		}
		clients[obj.Name] = client
	}
	return &Client{clients: clients}
}

func (c *Client) GetGRPCClient(name string) *grpc.ClientConn {
	return c.clients[name]
}

func (c *Client) Stop() {
	for _, client := range c.clients {
		_ = client.Close()
	}
}
