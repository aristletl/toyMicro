package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimiter struct {
	tokenChan chan struct{}
}

func NewTokenBucketLimiter(buffer int, interval time.Duration) TokenBucketLimiter {

	t := TokenBucketLimiter{
		tokenChan: make(chan struct{}, buffer),
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			t.tokenChan <- struct{}{}
		}
	}()
	return t
}

func (t *TokenBucketLimiter) BuildUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {

		select {
		case <-t.tokenChan:
			return handler(ctx, req)
		default:
			return nil, errors.New("限流")
		}
	}
}
