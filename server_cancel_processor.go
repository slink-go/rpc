package rpc

import (
	"context"
	"fmt"
	"sync"
)

type CancelProcessor struct {
	mu        sync.Mutex
	cancelers map[string]context.CancelFunc
}

func newRpcCancelProcessor() *CancelProcessor {
	return &CancelProcessor{
		cancelers: make(map[string]context.CancelFunc),
	}
}

func (s *CancelProcessor) Cancel(requestId string) (string, error) {
	s.mu.Lock()
	fn, ok := s.cancelers[requestId]
	if !ok {
		return "", nil
	}
	if ok {
		fn()
		delete(s.cancelers, requestId)
	}
	s.mu.Unlock()
	return fmt.Sprintf("request %s canceled", requestId), nil
}
func (s *CancelProcessor) Register(requestId string, cancel context.CancelFunc) {
	s.mu.Lock()
	_, ok := s.cancelers[requestId]
	if ok {
		delete(s.cancelers, requestId)
	}
	s.cancelers[requestId] = cancel
	s.mu.Unlock()
}
func (s *CancelProcessor) Unregister(requestId string) {
	s.mu.Lock()
	delete(s.cancelers, requestId)
	s.mu.Unlock()
}
