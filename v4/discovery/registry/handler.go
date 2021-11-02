package registry

import (
	"context"
	"time"

	pb "github.com/micro/micro/v3/proto/registry"
	mregistry "github.com/micro/micro/v3/service/registry"
	"github.com/micro/micro/v3/service/registry/util"
	"github.com/pydio/cells/v4/common/registry"
)

var reg = registry.Registry()

type Handler struct{}

func (h *Handler) GetService(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	resp := &pb.GetResponse{}

	ss, err := reg.GetService(req.GetService())
	if err != nil {
		return nil, err
	}

	var services []*pb.Service

	for _, s := range ss {
		services = append(services, util.ToProto(s))
	}

	resp.Services = services

	return resp, nil
}
func (h *Handler) Register(ctx context.Context, s *pb.Service) (*pb.EmptyResponse, error) {
	if err := reg.Register(util.ToService(s), mregistry.RegisterTTL(time.Duration(s.GetOptions().GetTtl())*time.Second)); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

func (h *Handler) Deregister(ctx context.Context, s *pb.Service) (*pb.EmptyResponse, error) {
	if err := reg.Deregister(util.ToService(s)); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

func (h *Handler) ListServices(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	resp := &pb.ListResponse{}

	ss, err := reg.ListServices()
	if err != nil {
		return nil, err
	}

	var services []*pb.Service
	for _, s := range ss {
		services = append(services, util.ToProto(s))
	}

	resp.Services = services

	return resp, nil
}

func (h *Handler) Watch(req *pb.WatchRequest, stream pb.Registry_WatchServer) error {
	var opts []mregistry.WatchOption
	if s := req.GetService(); s != "" {
		opts = append(opts, mregistry.WatchService(s))
	}

	w, err := reg.Watch(opts...)
	if err != nil {
		return err
	}

	for {
		res, err := w.Next()
		if err != nil {
			return err
		}

		stream.Send(&pb.Result{
			Action:  res.Action,
			Service: util.ToProto(res.Service),
		})
	}
}
