package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

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
	server     *grpc.Server
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

func (g *GRPC) Run(ctx context.Context, wg *sync.WaitGroup) {

	listen, err := net.Listen("tcp", g.gRPCConfig.Addr)
	if err != nil {
		panic(failure.Translate(err, appError.ErrServerStartFailed))
	}

	g.server = grpc.NewServer(grpc.UnaryInterceptor(g.accessLogInterceptor()))

	validateService := svcValidate.NewService(ctx, g.validator)
	pbValidate.RegisterValidateServiceServer(g.server, validateService)

	dslService := svcDSL.NewService(ctx, g.dslReader)
	pbDSL.RegisterDSLServiceServer(g.server, dslService)

	reflection.Register(g.server)

	go func() {
		if err := g.server.Serve(listen); err != nil {
			g.logger.Error(failure.Translate(err, appError.ErrServerInternalError).Error())
		}
	}()
	g.logger.Info("🟢 grpc server is started")

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
