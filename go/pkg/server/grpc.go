package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/authn"
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	svcDSL "github.com/shibukazu/open-ve/go/pkg/services/dsl/v1"
	svcHealth "github.com/shibukazu/open-ve/go/pkg/services/health/v1"
	svcSlave "github.com/shibukazu/open-ve/go/pkg/services/slave/v1"
	svcValidate "github.com/shibukazu/open-ve/go/pkg/services/validate/v1"
	"github.com/shibukazu/open-ve/go/pkg/slave"
	"github.com/shibukazu/open-ve/go/pkg/validator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbSlave "github.com/shibukazu/open-ve/go/proto/slave/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
	pbHealth "google.golang.org/grpc/health/grpc_health_v1"
)

type GRPC struct {
	mode           string
	dslReader      *reader.DSLReader
	validator      *validator.Validator
	slaveManager   *slave.SlaveManager
	slaveRegistrar *slave.SlaveRegistrar
	authenticator  authn.Authenticator
	gRPCConfig     *config.GRPCConfig
	logger         *slog.Logger
	server         *grpc.Server
}

func NewGrpc(
	mode string,
	gRPCConfig *config.GRPCConfig,
	logger *slog.Logger,
	validator *validator.Validator,
	dslReader *reader.DSLReader,
	slaveManager *slave.SlaveManager,
	slaveRegistrar *slave.SlaveRegistrar,
	authenticator authn.Authenticator,
) *GRPC {
	return &GRPC{
		mode:           mode,
		validator:      validator,
		dslReader:      dslReader,
		slaveManager:   slaveManager,
		slaveRegistrar: slaveRegistrar,
		authenticator:  authenticator,
		gRPCConfig:     gRPCConfig,
		logger:         logger,
	}
}

func (g *GRPC) Run(ctx context.Context, wg *sync.WaitGroup, mode string) {

	listen, err := net.Listen("tcp", ":"+g.gRPCConfig.Port)
	if err != nil {
		panic(failure.Translate(err, appError.ErrServerError, failure.Message("failed to listen")))
	}

	grpcServerOpts := []grpc.ServerOption{}
	grpcServerOpts = append(grpcServerOpts, grpc.ChainUnaryInterceptor([]grpc.UnaryServerInterceptor{
		g.accessLogInterceptor(),
		g.authnInterceptor(),
	}...))
	if g.gRPCConfig.TLS.Enabled {
		if g.gRPCConfig.TLS.CertPath == "" || g.gRPCConfig.TLS.KeyPath == "" {
			panic(failure.New(appError.ErrServerError, failure.Message("certPath and keyPath must be set")))
		}
		creds, err := credentials.NewServerTLSFromFile(g.gRPCConfig.TLS.CertPath, g.gRPCConfig.TLS.KeyPath)
		if err != nil {
			panic(failure.Translate(err, appError.ErrServerError, failure.Message("failed to load TLS credentials")))
		}
		grpcServerOpts = append(grpcServerOpts, grpc.Creds(creds))
	}

	g.server = grpc.NewServer(grpcServerOpts...)

	validateService := svcValidate.NewService(ctx, g.validator)
	pbValidate.RegisterValidateServiceServer(g.server, validateService)

	dslService := svcDSL.NewService(ctx, mode, g.dslReader, g.slaveRegistrar)
	pbDSL.RegisterDSLServiceServer(g.server, dslService)

	healthService := svcHealth.NewService(ctx)
	pbHealth.RegisterHealthServer(g.server, healthService)

	if mode == "master" {
		slaveService := svcSlave.NewService(ctx, g.slaveManager)
		pbSlave.RegisterSlaveServiceServer(g.server, slaveService)
	}

	reflection.Register(g.server)

	go func() {
		if err := g.server.Serve(listen); err != nil {
			g.logger.Error(failure.Translate(err, appError.ErrServerError, failure.Message("failed to serve grpc server")).Error())
		}
	}()

	if g.gRPCConfig.TLS.Enabled {
		g.logger.Info("🔒 grpc server: TLS is enabled")
	}
	g.logger.Info("🟢 grpc server: started")

	// graceful shutdown
	<-ctx.Done()

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	g.shutdown(ctxShutDown)
	wg.Done()
}

func (g *GRPC) shutdown(ctx context.Context) {
	ok := make(chan struct{})
	go func() {
		g.server.GracefulStop()
		close(ok)
	}()

	select {
	case <-ctx.Done():
		g.server.Stop()
		g.logger.Error("🛑 grpc server is stopped by timeout")
	case <-ok:
		g.logger.Info("🛑 grpc server is stopped")
	}
}

func (g *GRPC) accessLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		g.logger.Info("🔍 Access Log", slog.String("Method", info.FullMethod), slog.String("Request", fmt.Sprintf("%+v", req)))

		resp, err := handler(ctx, req)
		return resp, err
	}
}

func (g *GRPC) authnInterceptor() grpc.UnaryServerInterceptor {
	return grpcauth.UnaryServerInterceptor(
		func(ctx context.Context) (context.Context, error) {
			_, err := g.authenticator.Authenticate(ctx)
			if err != nil {
				return nil, err
			}
			return ctx, nil
		},
	)
}
