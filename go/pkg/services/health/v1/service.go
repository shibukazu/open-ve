package healthv1

import (
	"context"

	"google.golang.org/grpc/codes"
	pb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type Service struct {
	pb.UnimplementedHealthServer
}

func NewService(ctx context.Context) *Service {
	return &Service{}
}

func (h *Service) Check(context.Context, *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_SERVING,
	}, nil
}

func (h *Service) Watch(*pb.HealthCheckRequest, pb.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch is not implemented.")
}
