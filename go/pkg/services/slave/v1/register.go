package slavev1

import (
	"context"

	"github.com/shibukazu/open-ve/go/pkg/config"
	pb "github.com/shibukazu/open-ve/go/proto/slave/v1"
)

func (s *Service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	authnConfig := config.AuthnConfig{
		Method: req.Authn.GetMethod(),
		Preshared: config.PresharedConfig{
			Key: req.Authn.Preshared.GetKey(),
		},
	}

	s.slaveManager.RegisterSlave(
		req.Id,
		req.Address,
		req.TlsEnabled,
		req.ValidationIds,
		authnConfig,
	)
	return &pb.RegisterResponse{}, nil
}
