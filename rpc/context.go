package rpc

import "context"

type oneWayKey struct {
}

func CtxWithOneWay(ctx context.Context) context.Context {
	return context.WithValue(ctx, oneWayKey{}, true)
}

func IsOneWay(ctx context.Context) bool {
	return ctx.Value(oneWayKey{}) != nil
}
