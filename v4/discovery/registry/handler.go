package registry

import (
	"context"

	mregistry "github.com/micro/micro/v3/service/registry"
	pb "github.com/pydio/cells/v4/common/proto/registry"

	"github.com/pydio/cells/v4/common/registry"
)

type Handler struct{
	pb.UnimplementedRegistryServer

	reg registry.Registry
}

func (h *Handler) GetService(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	resp := &pb.GetResponse{}

	ss, err := h.reg.GetService(req.GetService())
	if err != nil {
		return nil, err
	}

	var services []*pb.Service

	for _, s := range ss {
		var p *mregistry.Service
		if ok := s.As(&p); ok {
			continue
		}
		services = append(services, registry.ToProto(p))
	}

	resp.Services = services

	return resp, nil
}
func (h *Handler) Register(ctx context.Context, s *pb.Service) (*pb.EmptyResponse, error) {
	if err := h.reg.Register(registry.ToService(s)); err != nil { // , mregistry.RegisterTTL(time.Duration(s.GetOptions().GetTtl())*time.Second)); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

func (h *Handler) Deregister(ctx context.Context, s *pb.Service) (*pb.EmptyResponse, error) {
	if err := h.reg.Deregister(registry.ToService(s)); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

func (h *Handler) ListServices(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	resp := &pb.ListResponse{}

	ss, err := h.reg.ListServices()
	if err != nil {
		return nil, err
	}

	var services []*pb.Service
	for _, s := range ss {
		var p *mregistry.Service
		if ok := s.As(&p); !ok {
			continue
		}
		services = append(services, registry.ToProto(p))
	}

	resp.Services = services

	return resp, nil
}

func (h *Handler) Watch(req *pb.WatchRequest, stream pb.Registry_WatchServer) error {
	return nil
	/*var opts []mregistry.WatchOption
	if s := req.GetService(); s != "" {
		opts = append(opts, mregistry.WatchService(s))
	}

	w, err := h.reg.Watch(opts...)
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
	}*/
}
