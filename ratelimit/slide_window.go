package ratelimit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	interval time.Duration
	max      int

	mutex sync.Mutex
	queue *list.List
}

func (s *SlideWindowLimiter) BuildUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		current := time.Now()
		windowStartTime := current.Add(-s.interval)
		s.mutex.Lock()
		reqTime := s.queue.Front()

		for reqTime != nil && reqTime.Value.(time.Time).Before(windowStartTime) {
			s.queue.Remove(reqTime)
			reqTime = s.queue.Front()
		}

		cnt := s.queue.Len()
		if cnt >= s.max {
			s.mutex.Unlock()
			err = errors.New("达到瓶颈")
			return
		}
		s.queue.PushBack(current)
		s.mutex.Unlock()
		return handler(ctx, req)
	}
}
