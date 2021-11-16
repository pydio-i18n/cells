package stubs

import (
	"context"
	"google.golang.org/grpc/metadata"
)

type StreamerStubCore struct {
	Ctx context.Context
}

func (s *StreamerStubCore) SetHeader(md metadata.MD) error {
	panic("implement me")
}

func (s *StreamerStubCore) SendHeader(md metadata.MD) error {
	panic("implement me")
}

func (s *StreamerStubCore) SetTrailer(md metadata.MD) {
	panic("implement me")
}

func (s *StreamerStubCore) Context() context.Context {
	return s.Ctx
}

func (s *StreamerStubCore) SendMsg(m interface{}) error {
	panic("implement me")
}

func (s *StreamerStubCore) RecvMsg(m interface{}) error {
	panic("implement me")
}
