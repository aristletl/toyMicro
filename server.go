package toy_micro

import (
	"context"
	"github.com/ztruane/toy-micro/registry"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerOption func(server *Server)

func WithRegistry(r registry.Registry) ServerOption {
	return func(s *Server) {
		s.r = r
	}
}

func WithServerWeight(weight uint32) ServerOption {
	return func(server *Server) {
		server.weight = weight
	}
}

func WithServerGroup(group string) ServerOption {
	return func(server *Server) {
		server.group = group
	}
}

type Server struct {
	name   string
	r      registry.Registry
	weight uint32
	group  string

	*grpc.Server
}

func NewServer(name string, opts ...ServerOption) *Server {
	s := &Server{
		name:   name,
		Server: grpc.NewServer(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 注册服务
	// 一定是先启动服务，再进行注册
	if s.r != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		err = s.r.Register(ctx, registry.ServiceInstance{
			ServiceName: s.name,
			Addr:        listener.Addr().String(),
			Weight:      s.weight,
			Group:       s.group,
		})
		cancel()
		if err != nil {
			return err
		}
	}

	return s.Serve(listener)
}
