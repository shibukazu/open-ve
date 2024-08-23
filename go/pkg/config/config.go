package config

type Config struct {
	Mode  string      `yaml:"mode"`
	Slave SlaveConfig `yaml:"slave"`
	Http  HttpConfig  `yaml:"http"`
	GRPC  GRPCConfig  `yaml:"grpc"`
	Store StoreConfig `yaml:"store"`
	Log   LogConfig   `yaml:"log"`
}

type SlaveConfig struct {
	Id                   string `yaml:"id"`
	SlaveGRPCAddr        string `yaml:"slaveAddr"`
	MasterGRPCTLSEnabled bool   `yaml:"masterGRPCTLSEnabled"`
	MasterGRPCAddr       string `yaml:"masterGRPCAddr"`
}

type HttpConfig struct {
	Port               string    `yaml:"port"`
	CORSAllowedOrigins []string  `yaml:"corsAllowedOrigins"`
	CORSAllowedHeaders []string  `yaml:"corsAllowedHeaders"`
	TLS                TLSConfig `yaml:"tls"`
}

type GRPCConfig struct {
	Port string    `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password" json:"-"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"poolSize"`
}

type StoreConfig struct {
	Engine string      `yaml:"engine"`
	Redis  RedisConfig `yaml:"redis"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertPath string `yaml:"certPath"`
	KeyPath  string `yaml:"keyPath"`
}

func DefaultConfig() *Config {
	return &Config{
		Mode: "master",
		Slave: SlaveConfig{
			Id:                   "",
			SlaveGRPCAddr:        "",
			MasterGRPCAddr:       "",
			MasterGRPCTLSEnabled: false,
		},
		Http: HttpConfig{
			Port:               "8080",
			CORSAllowedOrigins: []string{"*"},
			CORSAllowedHeaders: []string{"*"},
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		GRPC: GRPCConfig{
			Port: "9000",
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		Store: StoreConfig{
			Engine: "memory",
			Redis: RedisConfig{
				Addr:     "redis:6379",
				Password: "",
				DB:       0,
				PoolSize: 1000,
			},
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}
