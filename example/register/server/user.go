package main

import (
	"context"
	"fmt"
	"github.com/ztruane/toy-micro/example/register/gen"
)

type UserService struct {
	name string
	gen.UnimplementedUserServiceServer
}

func (u *UserService) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req.String())
	return &gen.GetByIdResp{
		User: &gen.User{
			Id:     12,
			Status: 1,
		},
	}, nil
}
