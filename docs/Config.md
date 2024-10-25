# Config

This document describes configuration of Open-VE.

You can configure Open-VE by CLI flags, environment variables, or config file.

## CLI Flags or Environment Variables

| CLI Args                             | Env                                        | Default      | Desc                                                                                   |
| ------------------------------------ | ------------------------------------------ | ------------ | -------------------------------------------------------------------------------------- |
| `--mode`                             | `OPEN-VE_MODE`                             | `master`     | master or slave                                                                        |
| `--slave-id`                         | `OPEN-VE_SLAVE_ID`                         |              | Unique slave ID (if mode is slave, this is required)                                   |
| `--slave-slave-http-addr`            | `OPEN-VE_SLAVE_SLAVE_HTTP_ADDR`            |              | HTTP server address (if mode is slave, this is required)                               |
| `--slave-master-http-addr`           | `OPEN-VE_SLAVE_MASTER_HTTP_ADDR`           |              | Master HTTP server address (if mode is slave, this is required)                        |
| `--slave-master-authn-method`        | `OPEN-VE_SLAVE_MASTER_AUTHN_METHOD`        | `none`       | Authentication method of the master server (preshared)                                 |
| `--slave-master-authn-preshared-key` | `OPEN-VE_SLAVE_MASTER_AUTHN_PRESHARED_KEY` |              | Preshared key of the master server (if authn method of the master server is preshared) |
| `--http-port`                        | `OPEN-VE_HTTP_PORT`                        | `8080`       | HTTP server port number                                                                |
| `--http-cors-allowed-origins`        | `OPEN-VE_HTTP_CORS_ALLOWED_ORIGINS`        | `["*"]`      | CORS allowed origins                                                                   |
| `--http-cors-allowed-headers`        | `OPEN-VE_HTTP_CORS_ALLOWED_HEADERS`        | `["*"]`      | CORS allowed headers                                                                   |
| `--http-tls-enabled`                 | `OPEN-VE_HTTP_TLS_ENABLED`                 | `false`      | HTTP server TLS enabled                                                                |
| `--http-tls-cert-path`               | `OPEN-VE_HTTP_TLS_CERT_PATH`               |              | HTTP server TLS cert path                                                              |
| `--http-tls-key-path`                | `OPEN-VE_HTTP_TLS_KEY_PATH`                |              | HTTP server TLS key path                                                               |
| `--grpc-port`                        | `OPEN-VE_GRPC_ADDR`                        | `9000`       | gRPC server port number                                                                |
| `--grpc-tls-enabled`                 | `OPEN-VE_GRPC_TLS_ENABLED`                 | `false`      | gRPC server TLS enabled                                                                |
| `--grpc-tls-cert-path`               | `OPEN-VE_GRPC_TLS_CERT_PATH`               |              | gRPC server TLS cert path                                                              |
| `--grpc-tls-key-path`                | `OPEN-VE_GRPC_TLS_KEY_PATH`                |              | gRPC server TLS key path                                                               |
| `--store-engine`                     | `OPEN-VE_STORE_ENGINE`                     | `memory`     | store engine (redis/memory)                                                            |
| `--store-redis-addr`                 | `OPEN-VE_STORE_REDIS_ADDR`                 | `redis:6379` | Redis address                                                                          |
| `--store-redis-password`             | `OPEN-VE_STORE_REDIS_PASSWORD`             |              | Redis password                                                                         |
| `--store-redis-db`                   | `OPEN-VE_STORE_REDIS_DB`                   | `0`          | Redis DB                                                                               |
| `--store-redis-pool-size`            | `OPEN-VE_STORE_REDIS_POOL_SIZE`            | `1000`       | Redis pool size                                                                        |
| `--log-level`                        | `OPEN-VE_LOG_LEVEL`                        | `info`       | Log level                                                                              |
| `--authn-method`                     | `OPEN-VE_AUTHN_METHOD`                     | `none`       | Authentication method of the server (preshared)                                        |
| `--authn-preshared-key`              | `OPEN-VE_AUTHN_PRESHARED_KEY`              |              | Preshared key of the server (if authn method is preshared)                             |

## Config File

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
