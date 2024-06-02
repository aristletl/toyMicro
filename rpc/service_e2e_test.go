package rpc

import (
	"context"
	"github.com/ztruane/toy-micro/rpc/serialize/json"
	"testing"
)

func TestServer_Start(t *testing.T) {
	s := NewServer()
	_ = s.Register(&UserServer{})
	s.MustRegister(json.Serializer{})
	s.Start(":8081")
}

type UserServer struct {
}

func (u *UserServer) Name() string {
	return "user-service"
}

func (u *UserServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	return &GetByIdResp{"Tom"}, nil
}
