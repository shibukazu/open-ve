package server

import (
	"context"
	"net"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	svcDSL "github.com/shibukazu/open-ve/go/pkg/services/dsl/v1"
	svcValidate "github.com/shibukazu/open-ve/go/pkg/services/validate/v1"
	"github.com/shibukazu/open-ve/go/pkg/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
)

type Grpc struct {
	dslReader *dsl.DSLReader
	validator *validator.Validator
}

func NewGrpc(validator *validator.Validator, dslReader *dsl.DSLReader) *Grpc {
	return &Grpc{
		validator: validator,
		dslReader: dslReader,
	}
}

func (g *Grpc) Run(ctx context.Context) {

	listen, err := net.Listen("tcp", grpcEndpoint)
	if err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed))
	}

	grpcServer := grpc.NewServer()

	validateService := svcValidate.NewService(ctx, g.validator)
	pbValidate.RegisterValidateServiceServer(grpcServer, validateService)

	dslService := svcDSL.NewService(ctx, g.dslReader)
	pbDSL.RegisterDSLServiceServer(grpcServer, dslService)

	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed))
	}
}


