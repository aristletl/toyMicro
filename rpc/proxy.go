package rpc

import (
	"context"
	"github.com/ztruane/toy-micro/rpc/message"
)

type Proxy interface {
	Invoke(ctx context.Context, req *message.Request) (*message.Response, error)
}

type Service interface {
	Name() string
}
