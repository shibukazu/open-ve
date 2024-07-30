package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-redis/redis"
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	"github.com/shibukazu/open-ve/go/pkg/logger"
	"github.com/shibukazu/open-ve/go/pkg/server"
	"github.com/shibukazu/open-ve/go/pkg/validator"
)

func main() {
	wg := &sync.WaitGroup{}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, os.Kill)
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
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		logger.Info("🚀 gateway server is starting...")
		gw.Run(ctx, wg)
	}(wg)

	grpc := server.NewGrpc(&cfg.GRPC, logger, validator, dslReader)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		logger.Info("🚀 grpc server is starting..")
		grpc.Run(ctx, wg)
	}(wg)

	wg.Wait()
	logger.Info("🛑 all server is stopped")
}
