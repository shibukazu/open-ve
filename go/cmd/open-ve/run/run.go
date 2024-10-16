package run

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-redis/redis"
	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"github.com/shibukazu/open-ve/go/pkg/authn"
	"github.com/shibukazu/open-ve/go/pkg/config"
	"github.com/shibukazu/open-ve/go/pkg/dsl/reader"
	"github.com/shibukazu/open-ve/go/pkg/logger"
	"github.com/shibukazu/open-ve/go/pkg/server"
	"github.com/shibukazu/open-ve/go/pkg/slave"
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			mode := viper.GetString("mode")
			if mode == "slave" {
				id := viper.GetString("slave.id")
				if id == "" {
					return failure.New(appError.ErrConfigError, failure.Message("ID of the slave server is required"))
				}
				slaveHTTPAddr := viper.GetString("slave.slaveHTTPAddr")
				if slaveHTTPAddr == "" {
					return failure.New(appError.ErrConfigError, failure.Message("HTTP address of the slave server is required"))
				}
				masterHTTPAddr := viper.GetString("slave.masterHTTPAddr")
				if masterHTTPAddr == "" {
					return failure.New(appError.ErrConfigError, failure.Message("HTTP address of the master server is required"))
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

	flags.String("slave-slave-http-addr", defaultConfig.Slave.SlaveHTTPAddr, "HTTP address of the slave server")
	MustBindPFlag("slave.slaveHTTPAddr", flags.Lookup("slave-slave-http-addr"))
	viper.MustBindEnv("slave.slaveHTTPAddr", "OPEN-VE_SLAVE_SLAVE_HTTP_ADDR")

	flags.String("slave-master-http-addr", defaultConfig.Slave.MasterHTTPAddr, "HTTP address of the master server")
	MustBindPFlag("slave.masterHTTPAddr", flags.Lookup("slave-master-http-addr"))
	viper.MustBindEnv("slave.masterHTTPAddr", "OPEN-VE_SLAVE_MASTER_HTTP_ADDR")

	flags.String("slave-master-authn-method", defaultConfig.Slave.MasterAuthn.Method, "Authentication method of the master server (preshared)")
	MustBindPFlag("slave.masterAuthn.method", flags.Lookup("slave-master-authn-method"))
	viper.MustBindEnv("slave.masterAuthn.method", "OPEN-VE_SLAVE_MASTER_AUTHN_METHOD")

	flags.String("slave-master-authn-preshared-key", defaultConfig.Slave.MasterAuthn.Preshared.Key, "Preshared key of the master server")
	MustBindPFlag("slave.masterAuthn.preshared.key", flags.Lookup("slave-master-authn-preshared-key"))
	viper.MustBindEnv("slave.masterAuthn.preshared.key", "OPEN-VE_SLAVE_MASTER_AUTHN_PRESHARED_KEY")

	// HTTP
	flags.String("http-port", defaultConfig.Http.Port, "HTTP server port")
	MustBindPFlag("http.port", flags.Lookup("http-port"))
	viper.MustBindEnv("http.port", "OPEN-VE_HTTP_PORT")

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
	flags.String("grpc-port", defaultConfig.GRPC.Port, "gRPC server port")
	MustBindPFlag("grpc.port", flags.Lookup("grpc-port"))
	viper.MustBindEnv("grpc.port", "OPEN-VE_GRPC_PORT")

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

	// Authn
	flags.String("authn-method", defaultConfig.Authn.Method, "Authentication method (preshared)")
	MustBindPFlag("authn.method", flags.Lookup("authn-method"))
	viper.MustBindEnv("authn.method", "OPEN-VE_AUTHN_METHOD")

	flags.String("authn-preshared-key", defaultConfig.Authn.Preshared.Key, "Preshared key")
	MustBindPFlag("authn.preshared.key", flags.Lookup("authn-preshared-key"))
	viper.MustBindEnv("authn.preshared.key", "OPEN-VE_AUTHN_PRESHARED_KEY")

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

	var nodeId string
	if cfg.Mode == "master" {
		nodeId = "master"
	} else {
		nodeId = cfg.Slave.Id
	}

	var store storePkg.Store
	switch cfg.Store.Engine {
	case "redis":
		redis := redis.NewClient(&redis.Options{
			Addr:     cfg.Store.Redis.Addr,
			Password: cfg.Store.Redis.Password,
			DB:       cfg.Store.Redis.DB,
			PoolSize: cfg.Store.Redis.PoolSize,
		})
		store = storePkg.NewRedisStore(nodeId, redis)
	case "memory":
		store = storePkg.NewMemoryStore(nodeId)
	default:
		panic("invalid store engine")
	}

	dslReader := reader.NewDSLReader(logger, store)
	validator := validator.NewValidator(logger, store)
	slaveManager := slave.NewSlaveManager(logger)
	slaveRegistrar := slave.NewSlaveRegistrar(cfg.Slave.Id, cfg.Slave.SlaveHTTPAddr, cfg.GRPC.TLS.Enabled, cfg.Authn, cfg.Slave.MasterHTTPAddr, cfg.Slave.MasterAuthn, dslReader, logger)
	var authenticator authn.Authenticator
	switch cfg.Authn.Method {
	case "preshared":
		logger.Info("🔐 authenticator: preshared key")
		authenticator = authn.NewPresharedKeyAuthenticator(cfg.Authn.Preshared.Key)
	default:
		logger.Warn("🔓 authenticator: none")
		authenticator = &authn.NoopAuthenticator{}
	}

	gw := server.NewGateway(cfg.Mode, &cfg.Http, &cfg.GRPC, logger, dslReader, slaveManager)
	wg.Add(1)

	logger.Info("🚀 Open-VE: starting...", slog.Any("config", cfg))
	go func(wg *sync.WaitGroup) {
		logger.Info("🚀 gateway server: starting...")
		gw.Run(ctx, wg)
	}(wg)

	grpc := server.NewGrpc(cfg.Mode, &cfg.GRPC, logger, validator, dslReader, slaveManager, slaveRegistrar, authenticator)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		logger.Info("🚀 grpc server: starting..")
		grpc.Run(ctx, wg, cfg.Mode)
	}(wg)

	if cfg.Mode == "slave" {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			logger.Info("🚀 slave registration timer: starting..")
			slaveRegistrar.RegisterTimer(ctx, wg)
		}(wg)
	}

	wg.Wait()
	logger.Info("🛑 all server and timer: stopped")
}
