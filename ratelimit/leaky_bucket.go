package ratelimit

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type LeakyBucketLimiter struct {
	producer *time.Ticker
	close    chan struct{}
}

func NewLeakyBucketLimiter(interval time.Duration) *LeakyBucketLimiter {
	res := &LeakyBucketLimiter{
		producer: time.NewTicker(interval),
		close:    make(chan struct{}),
	}

	return res

}

func (l *LeakyBucketLimiter) BuildUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-l.producer.C:
			return handler(ctx, req)
		}
	}
}

func (l *LeakyBucketLimiter) Close() error {
	l.producer.Stop()
	return nil
}
