package main

import (
	"context"

	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/server"
)

func main() {
	var ctx = context.Background()

	cfg := config.NewConfig()

	srv := server.NewGateway(&cfg.Server)
	srv.Run(ctx)
}
