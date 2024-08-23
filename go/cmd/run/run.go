package run

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	"github.com/shibukazu/open-ve/go/pkg/logger"
	"github.com/shibukazu/open-ve/go/pkg/server"
	"github.com/shibukazu/open-ve/go/pkg/slave"
	storePkg "github.com/shibukazu/open-ve/go/pkg/store"
	"github.com/shibukazu/open-ve/go/pkg/validator"
	pbSlave "github.com/shibukazu/open-ve/go/proto/slave/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the Open-VE server.",
		Long:  "Run the Open-VE server.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			mode := viper.GetString("mode")
			if mode == "slave" {
				id := viper.GetString("slave.id")
				if id == "" {
					return failure.New(appError.ErrConfigFileSyntaxError, failure.Message("ID of the slave server is required"))
				}
				masterAddr := viper.GetString("slave.masterGRPCAddr")
				if masterAddr == "" {
					return failure.New(appError.ErrConfigFileSyntaxError, failure.Message("gRPC address of the master server is required"))
				}
			}
			return nil
		},
		Run:  run,
		Args: cobra.NoArgs,
	}

	defaultConfig := config.DefaultConfig()

	flags := cmd.Flags()
	// Mode
	flags.String("mode", defaultConfig.Mode, "mode (master, slave)")
	MustBindPFlag("mode", flags.Lookup("mode"))
	viper.MustBindEnv("mode", "OPEN-VE_MODE")

	// Slave (If mode is slave, this is required)
	flags.String("slave-id", defaultConfig.Slave.Id, "ID of the slave server")
	MustBindPFlag("slave.id", flags.Lookup("slave-id"))
	viper.MustBindEnv("slave.id", "OPEN-VE_SLAVE_ID")

	flags.String("slave-master-grpc-addr", defaultConfig.Slave.MasterGRPCAddr, "gRPC address of the master server")
	MustBindPFlag("slave.masterGRPCAddr", flags.Lookup("slave-master-grpc-addr"))
	viper.MustBindEnv("slave.masterGRPCAddr", "OPEN-VE_SLAVE_MASTER_GRPC_ADDR")

	flags.Bool("slave-master-grpc-tls-enabled", defaultConfig.Slave.MasterGRPCTLSEnabled, "connect to master server with TLS")
	MustBindPFlag("slave.masterGRPCTLSEnabled", flags.Lookup("slave-master-grpc-tls-enabled"))
	viper.MustBindEnv("slave.masterGRPCTLSEnabled", "OPEN-VE_SLAVE_MASTER_GRPC_TLS_ENABLED")

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
	slaveManager := slave.NewSlaveManager(logger)

	gw := server.NewGateway(&cfg.Http, &cfg.GRPC, logger, dslReader)
	wg.Add(1)

	logger.Info("🚀 Open-VE: starting...", slog.Any("config", cfg))
	go func(wg *sync.WaitGroup) {
		logger.Info("🚀 gateway server: starting...")
		gw.Run(ctx, wg)
	}(wg)

	grpc := server.NewGrpc(&cfg.GRPC, logger, validator, dslReader, slaveManager)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		logger.Info("🚀 grpc server: starting..")
		grpc.Run(ctx, wg, cfg.Mode)
	}(wg)

	if cfg.Mode == "slave" {
		go func() {
			registerSlave(ctx, cfg, logger)
		}()
	}

	wg.Wait()
	logger.Info("🛑 all server: stopped")
}

func registerSlave(ctx context.Context, cfg config.Config, logger *slog.Logger) {
	var opts []grpc.DialOption

	if cfg.Slave.MasterGRPCTLSEnabled {
		creds := credentials.NewTLS(&tls.Config{})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))

	conn, err := grpc.Dial(cfg.Slave.MasterGRPCAddr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pbSlave.NewSlaveServiceClient(conn)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(5 * time.Second):
			_, err := client.Register(ctx, &pbSlave.RegisterRequest{
				Id:            cfg.Slave.Id,
				Address:       cfg.GRPC.Addr,
				ValidationIds: []string{"validation1", "validation2"},
			})
			if err != nil {
				logger.Error(failure.Translate(err, appError.ErrSlaveRegistrationFailed, failure.Message("Failed to register to master")).Error())
			} else {
				logger.Info(fmt.Sprintf("📓 slave (%s) registration success", cfg.Slave.Id))
			}
		}
	}
}
