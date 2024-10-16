package dslv1

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	dslPkg "github.com/shibukazu/open-ve/go/pkg/dsl"
	pb "github.com/shibukazu/open-ve/go/proto/dsl/v1"
)

func (s *Service) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	dsl, err := toDSL(req)
	if err != nil {
		s.logger.Error("failed to parse dsl: %v", slog.Any("code", failure.CodeOf(err)), slog.String("message", failure.MessageOf(err).String()), slog.String("details", fmt.Sprintf("%+v", err)))
		return nil, appError.ToGRPCError(err)
	}
	if err := s.dslReader.Register(ctx, dsl); err != nil {
		s.logger.Error("failed to register dsl: %v", slog.Any("code", failure.CodeOf(err)), slog.String("message", failure.MessageOf(err).String()), slog.String("details", fmt.Sprintf("%+v", err)))
		return nil, appError.ToGRPCError(err)
	}
	if s.mode == "slave" {
		s.slaveRegistrar.Register(ctx)
	}

	return &pb.RegisterResponse{}, nil
}

func toDSL(req *pb.RegisterRequest) (*dslPkg.DSL, error) {
	dsl := &dslPkg.DSL{}
	if req.Validations == nil {
		return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("validations is required"))
	}
	dsl.Validations = make([]dslPkg.Validation, len(req.Validations))
	for i, validation := range req.Validations {
		if validation.Id == "" {
			return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("id is required"))
		}
		if validation.Cels == nil {
			return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("cel is required"))
		}
		if validation.Variables == nil {
			return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("variables is required"))
		}
		dsl.Validations[i] = dslPkg.Validation{
			ID:        validation.Id,
			Cels:      validation.Cels,
			Variables: make([]dslPkg.Variable, len(validation.Variables)),
		}
		for j, variable := range validation.Variables {
			if variable.Name == "" {
				return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("variable name is required"))
			}
			if variable.Type == "" {
				return nil, failure.New(appError.ErrDSLSyntaxError, failure.Messagef("variable type is required"))
			}
			dsl.Validations[i].Variables[j] = dslPkg.Variable{
				Name: variable.Name,
				Type: variable.Type,
			}
		}
	}
	return dsl, nil
}
