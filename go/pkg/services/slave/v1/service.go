package slavev1

import (
	"context"
	"log/slog"

	"github.com/shibukazu/open-ve/go/pkg/slave"
	pb "github.com/shibukazu/open-ve/go/proto/slave/v1"
)

type Service struct {
	pb.UnimplementedSlaveServiceServer
	logger       *slog.Logger
	slaveManager *slave.SlaveManager
}

func NewService(ctx context.Context, logger *slog.Logger, slaveManager *slave.SlaveManager) *Service {
	return &Service{logger: logger, slaveManager: slaveManager}
}
