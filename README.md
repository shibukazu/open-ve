# Open-VE: Centralized and Consistent Data Validation Engine

![GitHub Release](https://img.shields.io/github/v/release/shibukazu/open-ve)
![GitHub License](https://img.shields.io/github/license/shibukazu/open-ve)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/shibukazu/open-ve)
![GitHub Repo stars](https://img.shields.io/github/stars/shibukazu/open-ve)

A powerful solution that simplifies the management of validation rules, ensuring consistent validation across all layers, including frontend, BFF, and microservices, through a single, simple API.

Open-VE offers an HTTP API and a gRPC API. We will provide a client SDK in the future.

> [!IMPORTANT]  
> This project is still under development and not ready for production use.

## Example

- [Example of Master Slave Architecture](docs/Master-Slave-Example.md)
- [Example of Monolithic Architecture](docs/Monolithic-Example.md)

## Getting Started

### Config

#### 1. CLI Flags or Environment Variables

| CLI Args                      | Env                                 | Default      | Desc                                                            |
| ----------------------------- | ----------------------------------- | ------------ | --------------------------------------------------------------- |
| `--mode`                      | `OPEN-VE_MODE`                      | `master`     | master or slave                                                 |
| `--slave-id`                  | `OPEN-VE_SLAVE_ID`                  |              | Unique slave ID (if mode is slave, this is required)            |
| `--slave-slave-http-addr`     | `OPEN-VE_SLAVE_SLAVE_HTTP_ADDR`     |              | HTTP server address (if mode is slave, this is required)        |
| `--slave-master-http-addr`    | `OPEN-VE_SLAVE_MASTER_HTTP_ADDR`    |              | Master HTTP server address (if mode is slave, this is required) |
| `--http-port`                 | `OPEN-VE_HTTP_PORT`                 | `8080`       | HTTP server port number                                         |
| `--http-cors-allowed-origins` | `OPEN-VE_HTTP_CORS_ALLOWED_ORIGINS` | `["*"]`      | CORS allowed origins                                            |
| `--http-cors-allowed-headers` | `OPEN-VE_HTTP_CORS_ALLOWED_HEADERS` | `["*"]`      | CORS allowed headers                                            |
| `--http-tls-enabled`          | `OPEN-VE_HTTP_TLS_ENABLED`          | `false`      | HTTP server TLS enabled                                         |
| `--http-tls-cert-path`        | `OPEN-VE_HTTP_TLS_CERT_PATH`        |              | HTTP server TLS cert path                                       |
| `--http-tls-key-path`         | `OPEN-VE_HTTP_TLS_KEY_PATH`         |              | HTTP server TLS key path                                        |
| `--grpc-port`                 | `OPEN-VE_GRPC_ADDR`                 | `9000`       | gRPC server port number                                         |
| `--grpc-tls-enabled`          | `OPEN-VE_GRPC_TLS_ENABLED`          | `false`      | gRPC server TLS enabled                                         |
| `--grpc-tls-cert-path`        | `OPEN-VE_GRPC_TLS_CERT_PATH`        |              | gRPC server TLS cert path                                       |
| `--grpc-tls-key-path`         | `OPEN-VE_GRPC_TLS_KEY_PATH`         |              | gRPC server TLS key path                                        |
| `--store-engine`              | `OPEN-VE_STORE_ENGINE`              | `memory`     | store engine (redis/memory)                                     |
| `--store-redis-addr`          | `OPEN-VE_STORE_REDIS_ADDR`          | `redis:6379` | Redis address                                                   |
| `--store-redis-password`      | `OPEN-VE_STORE_REDIS_PASSWORD`      |              | Redis password                                                  |
| `--store-redis-db`            | `OPEN-VE_STORE_REDIS_DB`            | `0`          | Redis DB                                                        |
| `--store-redis-pool-size`     | `OPEN-VE_STORE_REDIS_POOL_SIZE`     | `1000`       | Redis pool size                                                 |
| `--log-level`                 | `OPEN-VE_LOG_LEVEL`                 | `info`       | Log level                                                       |

#### 2. Config File

You can also use a config file in YAML format.

Place the `config.yaml` in the same directory or `$HOME/.open-ve/config.yaml`.

```yaml
mode: "master"
http:
  port: "8080"
  corsAllowedOrigins: ["*"]
  corsAllowedHeaders: ["*"]
  tls:
    enabled: false
    certPath: ""
    keyPath: ""
grpc:
  poer: "9000"
  tls:
    enabled: false
    certPath: ""
    keyPath: ""
store:
  engine: "redis" # redis or memory
  redis:
    addr: "redis:6379"
    password: ""
    db: 0
    poolSize: 1000
log:
  level: "info"
```

### Run

#### 1. brew

```bash
brew install shibukazu/tap/open-ve
open-ve run
```

#### 2. Build From Source

```bash
go build -o open-ve go/cmd/open-ve/main.go
./open-ve run
```

#### 3. Docker Compose

```bash
docker-compose up
```

## System Design

### Master-Slave Architecture

Open-VE supports a master-slave architecture designed for scalability and compatibility with microservice environments.

In slave mode, Open-VE connects to the master server and syncs validation rules every 30 seconds.

When a validation check request is made to the master server, it redirects the request across the connected slave servers.

Additionally, you can directly request validation checks to the slave servers.

![micro-validator](https://github.com/user-attachments/assets/e248d40c-bcc7-4219-a65a-5b243e101000)

### CEL

We use [CEL](https://github.com/google/cel-spec/blob/master/doc/langdef.md) as the expression language for validation rules.

Supported types:

| Type          | Support | Future Support |
| ------------- | ------- | -------------- |
| `int`         | ✅      |                |
| `uint`        | ✅      |                |
| `double`      | ✅      |                |
| `bool`        | ✅      |                |
| `string`      | ✅      |                |
| `bytes`       | ✅      |                |
| `list`        |         | ✅             |
| `map`         |         | ✅             |
| `null_type`   |         | ❓             |
| message names |         | ❓             |
| `type`        |         | ❓             |

## Documents

- [Try TLS Connection (on Master-Slave Architecture)](docs/Try-TLS_Connection.md)
