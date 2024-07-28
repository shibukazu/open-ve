package server

import (
	"context"
	"fmt"
	"log/slog"
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
	logger     *slog.Logger
}

func NewGrpc(
	gRPCConfig *config.GRPCConfig,
	logger *slog.Logger,
	validator *validator.Validator, dslReader *dsl.DSLReader) *GRPC {
	return &GRPC{
		validator:  validator,
		dslReader:  dslReader,
		gRPCConfig: gRPCConfig,
		logger:     logger,
	}
}

func (g *GRPC) Run(ctx context.Context) {

	listen, err := net.Listen("tcp", g.gRPCConfig.Addr)
	if err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed))
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(g.accessLogInterceptor()))

	validateService := svcValidate.NewService(ctx, g.validator)
	pbValidate.RegisterValidateServiceServer(grpcServer, validateService)

	dslService := svcDSL.NewService(ctx, g.dslReader)
	pbDSL.RegisterDSLServiceServer(grpcServer, dslService)

	reflection.Register(grpcServer)

	if err := grpcServer.Serve(listen); err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed))
	}
}

func (g *GRPC) accessLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		g.logger.Info("🔍 Access Log", slog.String("Method", info.FullMethod), slog.String("Request", fmt.Sprintf("%+v", req)))
		resp, err := handler(ctx, req)

		return resp, err
	}
}
