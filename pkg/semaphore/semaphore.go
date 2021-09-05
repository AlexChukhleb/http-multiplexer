package semaphore

import (
	"context"
	"errors"
	"time"
)

type Semaphore interface {
	Acquire(ctx context.Context, timeout time.Duration) error
	Release()
}

type semaphore struct {
	ch chan struct{}
}

func (s *semaphore) Acquire(ctx context.Context, timeout time.Duration) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(timeout):
		return errors.New("impossible to acquire")
	}
}

func (s *semaphore) Release() {
	select {
	case _ = <-s.ch:
		return
		//case <-time.After(time.Millisecond):
		//	return errors.New("need to Acquire first")
	}
}

func NewSemaphore(depth int) Semaphore {
	return &semaphore{
		ch: make(chan struct{}, depth),
	}
}
