package main

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	"github.com/shibukazu/open-ve/go/pkg/logger"
	"github.com/shibukazu/open-ve/go/pkg/server"
	"github.com/shibukazu/open-ve/go/pkg/validator"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	logger := logger.NewLogger(&cfg.Log)

	redis := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	dslReader := dsl.NewDSLReader(logger, redis)
	validator := validator.NewValidator(logger, redis)

	gw := server.NewGateway(&cfg.Http, &cfg.GRPC, logger, dslReader)
	go func() {
		logger.Info("🚀 gateway is running")
		gw.Run(ctx)
	}()

	grpc := server.NewGrpc(&cfg.GRPC, logger, validator, dslReader)
	go func() {
		logger.Info("🚀 grpc is running")
		grpc.Run(ctx)
	}()

	<-ctx.Done()
}
