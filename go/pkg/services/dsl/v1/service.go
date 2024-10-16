package dslv1

import (
	"context"
	"log/slog"

	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	"github.com/shibukazu/open-ve/go/pkg/slave"
	pb "github.com/shibukazu/open-ve/go/proto/dsl/v1"
)

type Service struct {
	pb.UnimplementedDSLServiceServer
	logger         *slog.Logger
	mode           string
	dslReader      *reader.DSLReader
	slaveRegistrar *slave.SlaveRegistrar
}

func NewService(ctx context.Context, logger *slog.Logger, mode string, dslReader *reader.DSLReader, slaveRegistrar *slave.SlaveRegistrar) *Service {
	return &Service{logger: logger, mode: mode, dslReader: dslReader, slaveRegistrar: slaveRegistrar}
}
