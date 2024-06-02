package boardcast

import (
	"context"
	"github.com/ztruane/toy-micro/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ClusterBuilder struct {
	r         registry.Registry
	serveName string
	opts      []grpc.DialOption
}

func (c *ClusterBuilder) BuildUnary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		instances, err := c.r.ListerServer(ctx, c.serveName)
		if err != nil {
			return err
		}

		var eg errgroup.Group
		for _, ins := range instances {
			instance := ins
			eg.Go(func() error {
				insCC, err := grpc.Dial(instance.Addr, c.opts...)
				if err != nil {
					return err
				}

				return invoker(ctx, method, req, reply, insCC, opts...)
			})
		}
		return eg.Wait()
	}
}
