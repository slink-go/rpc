package rpc

import (
	"context"
)

type CancelService struct {
	processor *CancelProcessor
}

func newRpcCancelService(processor *CancelProcessor) *CancelService {
	return &CancelService{
		processor: processor,
	}
}

func (s *CancelService) Cancel(_ context.Context, requestId string, response *string) (err error) {
	*response, err = s.processor.Cancel(requestId)
	return
}
