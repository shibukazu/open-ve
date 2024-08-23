package slavev1

import (
	"context"

	"github.com/shibukazu/open-ve/go/pkg/slave"
	pb "github.com/shibukazu/open-ve/go/proto/slave/v1"
)

type Service struct {
	pb.UnimplementedSlaveServiceServer
	slaveManager *slave.SlaveManager
}

func NewService(ctx context.Context, slaveManager *slave.SlaveManager) *Service {
	return &Service{slaveManager: slaveManager}
}
