package dslv1

import (
	"context"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	dslPkg "github.com/shibukazu/open-ve/go/pkg/dsl"
	pb "github.com/shibukazu/open-ve/go/proto/dsl/v1"
)



func (s *Service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	dsl, err := toDSL(req)
	if err != nil {
		return nil, appError.ToGRPCError(err)
	}
	if err := s.dslReader.Register(ctx, dsl); err != nil {
		return nil, appError.ToGRPCError(err)
	}

	return &pb.RegisterResponse{}, nil
}

func toDSL(req *pb.RegisterRequest) (*dslPkg.DSL, error) {
	dsl := &dslPkg.DSL{}
	if req.Validations == nil {
		return nil, failure.New(appError.ErrDSLServiceDSLSyntaxError, failure.Messagef("Validations is required"))
	}
	dsl.Validations = make([]dslPkg.Validation, len(req.Validations))
	for i, validation := range req.Validations {
		if validation.Id == "" {
			return nil, failure.New(appError.ErrDSLServiceDSLSyntaxError, failure.Messagef("Id is required"))
		}
		if validation.Cel == "" {
			return nil, failure.New(appError.ErrDSLServiceDSLSyntaxError, failure.Messagef("Cel is required"))
		}
		if validation.Variables == nil {
			return nil, failure.New(appError.ErrDSLServiceDSLSyntaxError, failure.Messagef("Variables is required"))
		}
		dsl.Validations[i] = dslPkg.Validation{
			ID:        validation.Id,
			Cel:       validation.Cel,
			Variables: make([]dslPkg.Variable, len(validation.Variables)),
		}
		for j, variable := range validation.Variables {
			if variable.Name == "" {
				return nil, failure.New(appError.ErrDSLServiceDSLSyntaxError, failure.Messagef("Variable Name is required"))
			}
			if variable.Type == "" {
				return nil, failure.New(appError.ErrDSLServiceDSLSyntaxError, failure.Messagef("Variable Type is required"))
			}
			dsl.Validations[i].Variables[j] = dslPkg.Variable{
				Name: variable.Name,
				Type: variable.Type,
			}
		}
	}
	return dsl, nil
}