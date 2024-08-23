package slavev1

import (
	"context"

	pb "github.com/shibukazu/open-ve/go/proto/slave/v1"
)

func (s *Service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.slaveManager.RegisterSlave(
		req.Id,
		req.Address,
		req.TlsEnabled,
		req.ValidationIds,
	)
	return &pb.RegisterResponse{}, nil
}
