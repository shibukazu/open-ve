package server

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pbDSL "github.com/shibukazu/open-ve/go/proto/dsl/v1"
	pbValidate "github.com/shibukazu/open-ve/go/proto/validate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {}

func NewGateway() *Gateway {
	return &Gateway{}
}

func (g *Gateway) Run(ctx context.Context) {
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

	if err := http.ListenAndServe(httpEndpoint, grpcGateway); err != nil {
		panic(err)
	}
}