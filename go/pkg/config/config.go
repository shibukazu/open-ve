package config

import (
	"log/slog"
	"os"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Http  HttpConfig  `yaml:"http"`
	GRPC  GRPCConfig  `yaml:"grpc"`
	Redis RedisConfig `yaml:"redis"`
	Log   LogConfig   `yaml:"log"`
}

type HttpConfig struct {
	Addr               string   `yaml:"addr"`
	CORSAllowedOrigins []string `yaml:"corsAllowedOrigins"`
	CORSAllowedHeaders []string `yaml:"corsAllowedHeaders"`
}

type GRPCConfig struct {
	Addr string `yaml:"addr"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password" json:"-"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"poolSize"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func NewConfig() *Config {
	config := defaultConfig()

	configPath := "config.yaml"

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Warn("config file not found, use default config")
			return config
		} else {
			panic(failure.Unexpected(err.Error()))
		}
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		panic(failure.Translate(err, appError.ErrConfigFileSyntaxError))
	}

	return config
}

func defaultConfig() *Config {
	return &Config{
		Http: HttpConfig{
			Addr:               "0.0.0.0:8080",
			CORSAllowedOrigins: []string{"*"},
			CORSAllowedHeaders: []string{"*"},
		},
		GRPC: GRPCConfig{
			Addr: "0.0.0.0:9000",
		},
		Redis: RedisConfig{
			Addr:     "redis:6379",
			Password: "",
			DB:       0,
			PoolSize: 1000,
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}
