package main

import (
	"context"
	toy_micro "github.com/ztruane/toy-micro"
	"github.com/ztruane/toy-micro/example/register/gen"
	"github.com/ztruane/toy-micro/loadbalance/roundrobin"
	"github.com/ztruane/toy-micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"log"
	"time"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2370"},
	})
	if err != nil {
		panic(err)
	}

	r, err := etcd.NewRegistry(cli)
	if err != nil {
		panic(err)
	}
	rb := toy_micro.NewResolverBuilder(r, 5*time.Second)

	pickerBuilder := &roundrobin.PickerBuilder{}
	builder := base.NewBalancerBuilder(pickerBuilder.Name(), pickerBuilder, base.Config{})
	balancer.Register(builder)

	cc, err := grpc.Dial("registry:///user-service",
		grpc.WithInsecure(),
		grpc.WithResolvers(rb),
		grpc.WithDefaultServiceConfig(`{"LoadBalancingPolicy":"ROUND_ROBIN"}`),
	)
	if err != nil {
		panic(err)
	}

	client := gen.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		resp, err := client.GetById(context.Background(), &gen.GetByIdReq{})
		if err != nil {
			panic(err)
		}
		log.Println(resp.User.String())
	}

}
