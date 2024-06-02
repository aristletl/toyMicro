package main

import (
	"context"
	"fmt"
	toy_micro "github.com/ztruane/toy-micro"
	"github.com/ztruane/toy-micro/example/register/gen"
	"github.com/ztruane/toy-micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	r, err := etcd.NewRegistry(&clientv3.Client{})
	if err != nil {
		panic(err)
	}
	builder := toy_micro.NewResolverBuilder(r, 5*time.Second)
	cc, err := grpc.Dial("registry:///user-service",
		grpc.WithInsecure(),
		grpc.WithResolvers(builder))
	if err != nil {
		panic(err)
	}

	client := gen.NewUserServiceClient(cc)
	resp, err := client.GetById(context.Background(), &gen.GetByIdReq{
		Id: 12,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.String())
}
