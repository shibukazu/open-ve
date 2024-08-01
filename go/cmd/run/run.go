package run

import (
	"context"
	"log/slog"
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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the Open-VE server.",
		Long:  "Run the Open-VE server.",
		Run:   run,
		Args:  cobra.NoArgs,
	}

	defaultConfig := config.DefaultConfig()

	flags := cmd.Flags()
	// HTTP
	flags.String("http-addr", defaultConfig.Http.Addr, "HTTP server address")
	viper.BindPFlag("http.addr", flags.Lookup("http-addr"))
	viper.BindEnv("http.addr", "OPEN-VE_HTTP_ADDR")

	flags.StringSlice("http-cors-allowed-origins", defaultConfig.Http.CORSAllowedOrigins, "CORS allowed origins")
	viper.BindPFlag("http.cors-allowed-origins", flags.Lookup("http-cors-allowed-origins"))
	viper.BindEnv("http.cors-allowed-origins", "OPEN-VE_HTTP_CORS_ALLOWED_ORIGINS")

	flags.StringSlice("http-cors-allowed-headers", defaultConfig.Http.CORSAllowedHeaders, "CORS allowed headers")
	viper.BindPFlag("http.cors-allowed-headers", flags.Lookup("http-cors-allowed-headers"))
	viper.BindEnv("http.cors-allowed-headers", "OPEN-VE_HTTP_CORS_ALLOWED_HEADERS")

	flags.Bool("http-tls-enabled", defaultConfig.Http.TLS.Enabled, "HTTP server TLS enabled")
	viper.BindPFlag("http.tls-enabled", flags.Lookup("http-tls-enabled"))
	viper.BindEnv("http.tls-enabled", "OPEN-VE_HTTP_TLS_ENABLED")

	flags.String("http-tls-cert-path", defaultConfig.Http.TLS.CertPath, "HTTP server TLS cert path")
	viper.BindPFlag("http.tls-cert-path", flags.Lookup("http-tls-cert-path"))
	viper.BindEnv("http.tls-cert-path", "OPEN-VE_HTTP_TLS_CERT_PATH")

	flags.String("http-tls-key-path", defaultConfig.Http.TLS.KeyPath, "HTTP server TLS key path")
	viper.BindPFlag("http.tls-key-path", flags.Lookup("http-tls-key-path"))
	viper.BindEnv("http.tls-key-path", "OPEN-VE_HTTP_TLS_KEY_PATH")
	// GRPC
	flags.String("grpc-addr", defaultConfig.GRPC.Addr, "gRPC server address")
	viper.BindPFlag("grpc.addr", flags.Lookup("grpc-addr"))
	viper.BindEnv("grpc.addr", "OPEN-VE_GRPC_ADDR")

	flags.Bool("grpc-tls-enabled", defaultConfig.GRPC.TLS.Enabled, "gRPC server TLS enabled")
	viper.BindPFlag("grpc.tls-enabled", flags.Lookup("grpc-tls-enabled"))
	viper.BindEnv("grpc.tls-enabled", "OPEN-VE_GRPC_TLS_ENABLED")

	flags.String("grpc-tls-cert-path", defaultConfig.GRPC.TLS.CertPath, "gRPC server TLS cert path")
	viper.BindPFlag("grpc.tls-cert-path", flags.Lookup("grpc-tls-cert-path"))
	viper.BindEnv("grpc.tls-cert-path", "OPEN-VE_GRPC_TLS_CERT_PATH")

	flags.String("grpc-tls-key-path", defaultConfig.GRPC.TLS.KeyPath, "gRPC server TLS key path")
	viper.BindPFlag("grpc.tls-key-path", flags.Lookup("grpc-tls-key-path"))
	viper.BindEnv("grpc.tls-key-path", "OPEN-VE_GRPC_TLS_KEY_PATH")
	// Redis
	flags.String("redis-addr", defaultConfig.Redis.Addr, "Redis address")
	viper.BindPFlag("redis.addr", flags.Lookup("redis-addr"))
	viper.BindEnv("redis.addr", "OPEN-VE_REDIS_ADDR")

	flags.String("redis-password", defaultConfig.Redis.Password, "Redis password")
	viper.BindPFlag("redis.password", flags.Lookup("redis-password"))
	viper.BindEnv("redis.password", "OPEN-VE_REDIS_PASSWORD")

	flags.Int("redis-db", defaultConfig.Redis.DB, "Redis DB")
	viper.BindPFlag("redis.db", flags.Lookup("redis-db"))
	viper.BindEnv("redis.db", "OPEN-VE_REDIS_DB")

	flags.Int("redis-pool-size", defaultConfig.Redis.PoolSize, "Redis pool size")
	viper.BindPFlag("redis.pool-size", flags.Lookup("redis-pool-size"))
	viper.BindEnv("redis.pool-size", "OPEN-VE_REDIS_POOL_SIZE")
	// Log
	flags.String("log-level", defaultConfig.Log.Level, "Log level")
	viper.BindPFlag("log.level", flags.Lookup("log-level"))
	viper.BindEnv("log.level", "OPEN-VE_LOG_LEVEL")

	return cmd
}

func run(cmd *cobra.Command, args []string) {
	configFileFound := true
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			configFileFound = false
		} else {
			panic(err)
		}
	}
	var cfg config.Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	logger := logger.NewLogger(&cfg.Log)

	if configFileFound {
		logger.Info("📖 config file found", slog.String("path", viper.ConfigFileUsed()))
	} else {
		logger.Info("📖 config file not found")
	}

	wg := &sync.WaitGroup{}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, os.Kill)
	defer cancel()

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

	logger.Info("🚀 Open-VE: starting...", slog.Any("config", cfg))
	go func(wg *sync.WaitGroup) {
		logger.Info("🚀 gateway server: starting...")
		gw.Run(ctx, wg)
	}(wg)

	grpc := server.NewGrpc(&cfg.GRPC, logger, validator, dslReader)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		logger.Info("🚀 grpc server: starting..")
		grpc.Run(ctx, wg)
	}(wg)

	wg.Wait()
	logger.Info("🛑 all server: stopped")
}
