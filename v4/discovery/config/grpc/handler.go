package grpc

import (
	"context"
	"fmt"
	pb "github.com/micro/micro/v3/proto/config"
	"github.com/pydio/cells/v4/x/configx"
)

var config = configx.New()

type Handler struct {
	serviceName string
}

func (h *Handler) ServiceName() string {
	return h.serviceName
}

func (h *Handler) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	return &pb.GetResponse{
		Value: &pb.Value{Data: config.Val().String()},
	}, nil
}

func (h *Handler) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse,error) {
	if err := config.Val(req.GetPath()).Set(req.GetValue()); err != nil {
		return nil, err
	}

	return &pb.SetResponse{}, nil
}

func (h *Handler) Delete(context.Context, *pb.DeleteRequest) (*pb.DeleteResponse,error) {
	return nil, fmt.Errorf("not implemented")
}

// These methods are here for backwards compatibility reasons
func (h *Handler) Read(context.Context, *pb.ReadRequest) (*pb.ReadResponse, error) {
	return nil, fmt.Errorf("not implemented")
}
