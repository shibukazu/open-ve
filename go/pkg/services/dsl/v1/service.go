package dslv1

import (
	"context"

	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	pb "github.com/shibukazu/open-ve/go/proto/dsl/v1"
)

type Service struct {
	pb.UnimplementedDSLServiceServer
	dslReader *reader.DSLReader
}

func NewService(ctx context.Context, dslReader *reader.DSLReader) *Service {
	return &Service{dslReader: dslReader}
}
