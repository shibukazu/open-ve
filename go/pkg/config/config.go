package config

import (
	"os"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Redis  RedisConfig  `yaml:"redis"`
}

type ServerConfig struct {
	Grpc GrpcConfig `yaml:"grpc"`
	Http HttpConfig `yaml:"http"`
}

type GrpcConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type HttpConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"poolSize"`
}

func NewConfig() *Config {
	config := defaultConfig()

	configPath := "config.yaml"

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// TODO warning
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

func defaultConfig () *Config {
	return &Config{
		Server: ServerConfig{
			Grpc: GrpcConfig{
				Host: "localhost",
				Port: "9000",
			},
			Http: HttpConfig{
				Host: "localhost",
				Port: "8080",
			},
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
			PoolSize: 1000,
		},
	}
}