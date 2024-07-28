package server

import (
	"context"
	"net"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	svcDSL "github.com/shibukazu/open-ve/go/pkg/services/dsl/v1"
	svcValidate "github.com/shibukazu/open-ve/go/pkg/services/validate/v1"
	"github.com/shibukazu/open-ve/go/pkg/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
)

type GRPC struct {
	dslReader  *dsl.DSLReader
	validator  *validator.Validator
	gRPCConfig *config.GRPCConfig
}

func NewGrpc(
	gRPCConfig *config.GRPCConfig,
	validator *validator.Validator, dslReader *dsl.DSLReader) *GRPC {
	return &GRPC{
		validator:  validator,
		dslReader:  dslReader,
		gRPCConfig: gRPCConfig,
	}
}

func (g *GRPC) Run(ctx context.Context) {

	listen, err := net.Listen("tcp", g.gRPCConfig.Addr)
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
