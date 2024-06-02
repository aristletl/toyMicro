package ratelimit

import (
	"context"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

type FixedWindowLimiter struct {
	interval  int64
	timestamp int64

	max int64
	cnt int64
}

//
//func (f *FixedWindowLimiter) OnRejection(ctx context.Context, req any) (any, error) {
//	return f.
//}

func (f *FixedWindowLimiter) BuildUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		current := time.Now().Unix()
		timestamp := atomic.LoadInt64(&f.timestamp)
		if timestamp+f.interval < current {
			if atomic.CompareAndSwapInt64(&f.timestamp, timestamp, current) {
				atomic.StoreInt64(&f.cnt, 0)
			}
		}

		cnt := atomic.AddInt64(&f.cnt, 1)
		if cnt > f.max {
			//return f.
		}
		return handler(ctx, req)
	}
}
