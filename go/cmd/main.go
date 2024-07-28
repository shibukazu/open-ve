package main

import (
	"context"
	"log/slog"

	"github.com/go-redis/redis"
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	"github.com/shibukazu/open-ve/go/pkg/server"
	"github.com/shibukazu/open-ve/go/pkg/validator"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	redis := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})
	dslReader := dsl.NewDSLReader(redis)
	validator := validator.NewValidator(redis)
	gw := server.NewGateway(&cfg.Http, &cfg.GRPC)
	go func() {
		slog.Info("🚀gateway is running")
		gw.Run(ctx)
	}()

	grpc := server.NewGrpc(&cfg.GRPC, validator, dslReader)
	go func() {
		slog.Info("🚀grpc is running")
		grpc.Run(ctx)
	}()

	<-ctx.Done()
}
