package dslv1

import (
	"context"

	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	"github.com/shibukazu/open-ve/go/pkg/slave"
	pb "github.com/shibukazu/open-ve/go/proto/dsl/v1"
)

type Service struct {
	pb.UnimplementedDSLServiceServer
	mode           string
	dslReader      *reader.DSLReader
	slaveRegistrar *slave.SlaveRegistrar
}

func NewService(ctx context.Context, mode string, dslReader *reader.DSLReader, slaveRegistrar *slave.SlaveRegistrar) *Service {
	return &Service{mode: mode, dslReader: dslReader, slaveRegistrar: slaveRegistrar}
}
