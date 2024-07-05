package server

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/shibukazu/open-ve/go/pkg/config"
	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	serverConfig *config.ServerConfig
}

func NewGateway(serverConfig *config.ServerConfig) *Gateway {
	return &Gateway{
		serverConfig: serverConfig,
	}
}

func (g *Gateway) Run(ctx context.Context) {
	grpcEndpoint := g.serverConfig.Grpc.Host + ":" + g.serverConfig.Grpc.Port
	grpcGateway := runtime.NewServeMux()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := pbValidate.RegisterValidateServiceHandlerFromEndpoint(ctx, grpcGateway, grpcEndpoint, opts); err != nil {
		panic(err)
	}
	if err := pbDSL.RegisterDSLServiceHandlerFromEndpoint(ctx, grpcGateway, grpcEndpoint, opts); err != nil {
		panic(err)
	}

	if err := http.ListenAndServe(":"+g.serverConfig.Http.Port, grpcGateway); err != nil {
		panic(err)
	}
}