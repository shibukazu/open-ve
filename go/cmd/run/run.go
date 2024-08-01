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
	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	"github.com/shibukazu/open-ve/go/pkg/logger"
	"github.com/shibukazu/open-ve/go/pkg/server"
	storePkg "github.com/shibukazu/open-ve/go/pkg/store"
	"github.com/shibukazu/open-ve/go/pkg/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	MustBindPFlag("http.addr", flags.Lookup("http-addr"))
	viper.MustBindEnv("http.addr", "OPEN-VE_HTTP_ADDR")

	flags.StringSlice("http-cors-allowed-origins", defaultConfig.Http.CORSAllowedOrigins, "CORS allowed origins")
	MustBindPFlag("http.corsAllowedOrigins", flags.Lookup("http-cors-allowed-origins"))
	viper.MustBindEnv("http.corsAllowedOrigins", "OPEN-VE_HTTP_CORS_ALLOWED_ORIGINS")

	flags.StringSlice("http-cors-allowed-headers", defaultConfig.Http.CORSAllowedHeaders, "CORS allowed headers")
	MustBindPFlag("http.corsAllowedHeaders", flags.Lookup("http-cors-allowed-headers"))
	viper.MustBindEnv("http.corsAllowedHeaders", "OPEN-VE_HTTP_CORS_ALLOWED_HEADERS")

	flags.Bool("http-tls-enabled", defaultConfig.Http.TLS.Enabled, "HTTP server TLS enabled")
	MustBindPFlag("http.tls.enabled", flags.Lookup("http-tls-enabled"))
	viper.MustBindEnv("http.tls.enabled", "OPEN-VE_HTTP_TLS_ENABLED")

	flags.String("http-tls-cert-path", defaultConfig.Http.TLS.CertPath, "HTTP server TLS cert path")
	MustBindPFlag("http.tls.certPath", flags.Lookup("http-tls-cert-path"))
	viper.MustBindEnv("http.tls.certPath", "OPEN-VE_HTTP_TLS_CERT_PATH")

	flags.String("http-tls-key-path", defaultConfig.Http.TLS.KeyPath, "HTTP server TLS key path")
	MustBindPFlag("http.tls.keyPath", flags.Lookup("http-tls-key-path"))
	viper.MustBindEnv("http.tls.keyPath", "OPEN-VE_HTTP_TLS_KEY_PATH")
	// GRPC
	flags.String("grpc-addr", defaultConfig.GRPC.Addr, "gRPC server address")
	MustBindPFlag("grpc.addr", flags.Lookup("grpc-addr"))
	viper.MustBindEnv("grpc.addr", "OPEN-VE_GRPC_ADDR")

	flags.Bool("grpc-tls-enabled", defaultConfig.GRPC.TLS.Enabled, "gRPC server TLS enabled")
	MustBindPFlag("grpc.tls.enabled", flags.Lookup("grpc-tls-enabled"))
	viper.MustBindEnv("grpc.tls.enabled", "OPEN-VE_GRPC_TLS_ENABLED")

	flags.String("grpc-tls-cert-path", defaultConfig.GRPC.TLS.CertPath, "gRPC server TLS cert path")
	MustBindPFlag("grpc.tls.certPath", flags.Lookup("grpc-tls-cert-path"))
	viper.MustBindEnv("grpc.tls.certPath", "OPEN-VE_GRPC_TLS_CERT_PATH")

	flags.String("grpc-tls-key-path", defaultConfig.GRPC.TLS.KeyPath, "gRPC server TLS key path")
	MustBindPFlag("grpc.tls.keyPath", flags.Lookup("grpc-tls-key-path"))
	viper.MustBindEnv("grpc.tls.keyPath", "OPEN-VE_GRPC_TLS_KEY_PATH")
	// Store
	flags.String("store-engine", defaultConfig.Store.Engine, "store engine (memory, redis)")
	MustBindPFlag("store.engine", flags.Lookup("store-engine"))
	viper.MustBindEnv("store.engine", "OPEN-VE_STORE_ENGINE")

	flags.String("store-redis-addr", defaultConfig.Store.Redis.Addr, "Redis address")
	MustBindPFlag("store.redis.addr", flags.Lookup("store-redis-addr"))
	viper.MustBindEnv("store.redis.addr", "OPEN-VE_STORE_REDIS_ADDR")

	flags.String("store-redis-password", defaultConfig.Store.Redis.Password, "Redis password")
	MustBindPFlag("store.redis.password", flags.Lookup("store-redis-password"))
	viper.MustBindEnv("store.redis.password", "OPEN-VE_STORE_REDIS_PASSWORD")

	flags.Int("store-redis-db", defaultConfig.Store.Redis.DB, "Redis DB")
	MustBindPFlag("store.redis.db", flags.Lookup("store-redis-db"))
	viper.MustBindEnv("store.redis.db", "OPEN-VE_STORE_REDIS_DB")

	flags.Int("store-redis-pool-size", defaultConfig.Store.Redis.PoolSize, "Redis pool size")
	MustBindPFlag("store.redis.poolSize", flags.Lookup("store-redis-pool-size"))
	viper.MustBindEnv("store.redis.poolSize", "OPEN-VE_STORE_REDIS_POOL_SIZE")
	// Log
	flags.String("log-level", defaultConfig.Log.Level, "Log level")
	MustBindPFlag("log.level", flags.Lookup("log-level"))
	viper.MustBindEnv("log.level", "OPEN-VE_LOG_LEVEL")

	return cmd
}

func MustBindPFlag(key string, flag *pflag.Flag) {
	if err := viper.BindPFlag(key, flag); err != nil {
		panic(err)
	}
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

	var store storePkg.Store
	switch cfg.Store.Engine {
	case "redis":
		redis := redis.NewClient(&redis.Options{
			Addr:     cfg.Store.Redis.Addr,
			Password: cfg.Store.Redis.Password,
			DB:       cfg.Store.Redis.DB,
			PoolSize: cfg.Store.Redis.PoolSize,
		})
		store = storePkg.NewRedisStore(redis)
	case "memory":
		store = storePkg.NewMemoryStore()
	default:
		panic("invalid store engine")
	}

	dslReader := reader.NewDSLReader(logger, store)
	validator := validator.NewValidator(logger, store)

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
