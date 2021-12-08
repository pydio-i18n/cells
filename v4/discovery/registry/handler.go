package registry

import (
	"context"

	"github.com/pydio/cells/v4/common"

	pb "github.com/pydio/cells/v4/common/proto/registry"

	"github.com/pydio/cells/v4/common/registry"
	"github.com/pydio/cells/v4/common/registry/service"
)

type Handler struct {
	pb.UnimplementedRegistryServer

	reg registry.Registry
}

func (h *Handler) Name() string {
	return common.ServiceGrpcNamespace_ + common.ServiceRegistry
}

func (h *Handler) StartService(ctx context.Context, req *pb.StartServiceRequest) (*pb.EmptyResponse, error) {
	if err := h.reg.StartService(req.Service); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}
func (h *Handler) StopService(ctx context.Context, req *pb.StopServiceRequest) (*pb.EmptyResponse, error) {
	if err := h.reg.StopService(req.Service); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}
func (h *Handler) GetService(ctx context.Context, req *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	resp := &pb.GetServiceResponse{}

	ss, err := h.reg.GetService(req.GetService())
	if err != nil {
		return nil, err
	}

	resp.Service = service.ToProtoService(ss)

	return resp, nil
}
func (h *Handler) RegisterService(ctx context.Context, s *pb.Service) (*pb.EmptyResponse, error) {
	if err := h.reg.RegisterService(service.ToService(s)); err != nil { // , mregistry.RegisterTTL(time.Duration(s.GetOptions().GetTtl())*time.Second)); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

func (h *Handler) DeregisterService(ctx context.Context, s *pb.Service) (*pb.EmptyResponse, error) {
	if err := h.reg.DeregisterService(service.ToService(s)); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

func (h *Handler) ListServices(ctx context.Context, req *pb.ListServicesRequest) (*pb.ListServicesResponse, error) {
	resp := &pb.ListServicesResponse{}

	ss, err := h.reg.ListServices()
	if err != nil {
		return nil, err
	}

	var services []*pb.Service
	for _, s := range ss {
		services = append(services, service.ToProtoService(s))
	}

	resp.Services = services

	return resp, nil
}

func (h *Handler) WatchServices(req *pb.WatchServicesRequest, stream pb.Registry_WatchServicesServer) error {

	//TODO v4 options
	//var opts []registry.WatchOption
	//if s := req.GetService(); s != "" {
	//	opts = append(opts, registry.WatchService(s))
	//}

	w, err := h.reg.WatchServices()
	if err != nil {
		return err
	}

	for {
		res, err := w.Next()
		if err != nil {
			return err
		}

		stream.Send(&pb.Result{
			Action:  res.Action(),
			Service: service.ToProtoService(res.Service()),
		})
	}
}
