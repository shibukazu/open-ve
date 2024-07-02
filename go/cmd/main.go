package main

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	"github.com/shibukazu/open-ve/go/pkg/server"
	"github.com/shibukazu/open-ve/go/pkg/validator"
)

func main() {
	var ctx = context.Background()

	cfg := config.NewConfig()

	redis := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:      cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	dslReader := dsl.NewDSLReader(redis)
	validator := validator.NewValidator(redis)

	srv := server.NewServer(validator, dslReader, &cfg.Server)
	srv.Run(ctx)
}
