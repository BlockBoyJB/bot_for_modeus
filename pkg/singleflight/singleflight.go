package singleflight

import (
	"context"
	"fmt"
	"sync"
)

type call struct {
	value any
	err   error
	sem   chan struct{}
}

type Flight struct {
	mx    sync.Mutex
	calls map[string]*call
}

func NewFlight() *Flight {
	return &Flight{
		calls: make(map[string]*call),
	}
}

func (s *Flight) Do(ctx context.Context, key string, action func() (any, error)) (any, error) {
	s.mx.Lock()
	if c, ok := s.calls[key]; ok {
		s.mx.Unlock()
		return s.wait(ctx, c)
	}
	c := &call{
		sem: make(chan struct{}),
	}
	s.calls[key] = c
	s.mx.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.err = fmt.Errorf("[PANIC] single flight %v", r)
			}
			close(c.sem)

			s.mx.Lock()
			delete(s.calls, key)
			s.mx.Unlock()
		}()
		c.value, c.err = action()
	}()

	return s.wait(ctx, c)
}

func (s *Flight) wait(ctx context.Context, call *call) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-call.sem:
		return call.value, call.err
	}
}
