package main

import (
	toy_micro "github.com/ztruane/toy-micro"
	"github.com/ztruane/toy-micro/example/register/gen"
	"github.com/ztruane/toy-micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2370"},
	})
	if err != nil {
		panic(err)
	}

	r, err := etcd.NewRegistry(client)
	if err != nil {
		panic(err)
	}

	serve := toy_micro.NewServer("user-service", toy_micro.WithRegistry(r))
	us := &UserService{
		name: "abc",
	}
	gen.RegisterUserServiceServer(serve, us)
	serve.Start(":8081")
}
