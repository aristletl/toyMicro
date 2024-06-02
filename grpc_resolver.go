package toy_micro

import (
	"context"
	"fmt"
	"github.com/ztruane/toy-micro/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

type grpcResolverBuilder struct {
	r       registry.Registry
	timeout time.Duration
}

func NewResolverBuilder(r registry.Registry, timeout time.Duration) resolver.Builder {
	return &grpcResolverBuilder{
		r:       r,
		timeout: timeout,
	}
}

func (g *grpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	res := &grpcResolver{
		target: target,
		cc:     cc,
		r:      g.r,
		close:  make(chan struct{}),

		timeout: g.timeout,
	}
	res.resolve()

	return res, res.watch()
}

func (g *grpcResolverBuilder) Scheme() string {
	return "registry"
}

type grpcResolver struct {
	r      registry.Registry
	target resolver.Target
	cc     resolver.ClientConn
	close  chan struct{}

	timeout time.Duration
}

// ResolveNow 解析服务——立即执行服务发现——去问一下注册中心
func (g *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.resolve()
}

func (g *grpcResolver) resolve() {
	// 可用服务实例列表，考虑到设置超时，因此需要创建一个超时的ctx
	// 这个超时时间是由Builder透传过来的
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	instances, err := g.r.ListerServer(ctx, g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return
	}

	address := make([]resolver.Address, 0, len(instances))
	for _, ins := range instances {
		address = append(address, resolver.Address{
			//ServerName: ins.na
			// 定位信息，ip+端口
			Addr:       ins.Addr,
			Attributes: attributes.New("weight", ins.Weight).WithValue("group", ins.Group),
		})
	}

	// 用来告诉grpc，该服务可用的实例列表
	err = g.cc.UpdateState(resolver.State{Addresses: address})
	if err != nil {
		g.cc.ReportError(err)
	}
}

func (g *grpcResolver) watch() error {
	ech, err := g.r.Subscribe(context.Background(), g.target.Endpoint())
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event := <-ech:

				//switch event.Type {
				//case registry.EventTypeAdd:
				//case registry.EventTypeDelete:
				//case registry.EventTypeUnknown:
				//case registry.EventTypeUpdate:
				//}
				g.resolve()
				fmt.Print(event)
			case <-g.close:
				close(g.close)
				return
			}
		}
	}()

	return nil
}

func (g *grpcResolver) Close() {
	g.close <- struct{}{}
}
