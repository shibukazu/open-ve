package server

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/shibukazu/open-ve/go/pkg/config"
	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	httpConfig *config.HttpConfig
	gRPCConfig *config.GRPCConfig
}

func NewGateway(
	httpConfig *config.HttpConfig,
	gRPCConfig *config.GRPCConfig,
) *Gateway {
	return &Gateway{
		httpConfig: httpConfig,
		gRPCConfig: gRPCConfig,
	}
}

func (g *Gateway) Run(ctx context.Context) {
	grpcGateway := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := pbValidate.RegisterValidateServiceHandlerFromEndpoint(ctx, grpcGateway, g.gRPCConfig.Addr, opts); err != nil {
		panic(err)
	}

	if err := pbDSL.RegisterDSLServiceHandlerFromEndpoint(ctx, grpcGateway, g.gRPCConfig.Addr, opts); err != nil {
		panic(err)
	}

	withCors := cors.New(cors.Options{
		AllowedOrigins:   g.httpConfig.CORSAllowedOrigins,
		AllowedHeaders:   g.httpConfig.CORSAllowedHeaders,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(grpcGateway)

	if err := http.ListenAndServe(g.httpConfig.Addr, withCors); err != nil {
		panic(err)
	}
}
