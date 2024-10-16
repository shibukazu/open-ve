package dslv1

import (
	"context"

	"github.com/shibukazu/open-ve/go/pkg/appError"
	dslPkg "github.com/shibukazu/open-ve/go/pkg/dsl"
	"github.com/shibukazu/open-ve/go/pkg/logger"
	pb "github.com/shibukazu/open-ve/go/proto/dsl/v1"
)

func (s *Service) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	dsl, err := s.dslReader.Read(ctx)
	if err != nil {
		logger.LogError(s.logger, err)
		return nil, appError.ToGRPCError(err)
	}
	return toProto(dsl), nil
}

func toProto(dsl *dslPkg.DSL) *pb.ReadResponse {
	res := &pb.ReadResponse{}
	res.Validations = make([]*pb.Validation, len(dsl.Validations))
	for i, validation := range dsl.Validations {
		res.Validations[i] = &pb.Validation{
			Id:        validation.ID,
			Cels:      validation.Cels,
			Variables: make([]*pb.Variable, len(validation.Variables)),
		}
		for j, variable := range validation.Variables {
			res.Validations[i].Variables[j] = &pb.Variable{
				Name: variable.Name,
				Type: variable.Type,
			}
		}
	}
	return res
}
