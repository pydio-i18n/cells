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

type ClientServerStreamerCore struct {
	Ctx         context.Context
	SendHandler func(interface{}) error
	RespChan    chan interface{}
	closed      bool
	header      metadata.MD
	trailer     metadata.MD
}

func (cs *ClientServerStreamerCore) SetHeader(md metadata.MD) error {
	cs.header = md
	return nil
}

func (cs *ClientServerStreamerCore) SendHeader(md metadata.MD) error {
	return nil
}

func (cs *ClientServerStreamerCore) SetTrailer(md metadata.MD) {
	cs.trailer = md
}

func (cs *ClientServerStreamerCore) Header() (metadata.MD, error) {
	return cs.header, nil
}

func (cs *ClientServerStreamerCore) Trailer() metadata.MD {
	return cs.trailer
}

func (cs *ClientServerStreamerCore) CloseSend() error {
	if cs.closed {
		return nil
	}
	close(cs.RespChan)
	cs.closed = true
	return nil
}

func (cs *ClientServerStreamerCore) Context() context.Context {
	return cs.Ctx
}

func (cs *ClientServerStreamerCore) SendMsg(m interface{}) error {
	return cs.SendHandler(m)
}
