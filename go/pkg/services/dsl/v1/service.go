package dslv1

import (
	"context"

	"github.com/shibukazu/open-ve/go/pkg/dsl"
	pb "github.com/shibukazu/open-ve/go/proto/dsl/v1"
)

type Service struct {
	pb.UnimplementedDSLServiceServer
	dslReader *dsl.DSLReader
}

func NewService(ctx context.Context, dslReader *dsl.DSLReader) *Service {
	return &Service{dslReader: dslReader}
}