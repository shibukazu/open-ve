# Open-VE: Centralized and Consistent Data Validation Engine

![GitHub Release](https://img.shields.io/github/v/release/shibukazu/open-ve)
![GitHub License](https://img.shields.io/github/license/shibukazu/open-ve)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/shibukazu/open-ve)
![GitHub Repo stars](https://img.shields.io/github/stars/shibukazu/open-ve)

A powerful solution that simplifies the management of validation rules, ensuring consistent validation across all layers, including frontend, BFF, and microservices, through a single, simple API.

Open-VE offers an HTTP API and a gRPC API. We will provide a client SDK in the future.

> [!IMPORTANT]  
> This project is still under development and not ready for production use.

## Getting Started

### Config

#### 1. CLI Flags or Environment Variables

| CLI Args                      | Env                                 | Default      | Desc                        |
| ----------------------------- | ----------------------------------- | ------------ | --------------------------- |
| `--http-addr`                 | `OPEN-VE_HTTP_ADDR`                 | `:8080`      | HTTP server address         |
| `--http-cors-allowed-origins` | `OPEN-VE_HTTP_CORS_ALLOWED_ORIGINS` | `["*"]`      | CORS allowed origins        |
| `--http-cors-allowed-headers` | `OPEN-VE_HTTP_CORS_ALLOWED_HEADERS` | `["*"]`      | CORS allowed headers        |
| `--http-tls-enabled`          | `OPEN-VE_HTTP_TLS_ENABLED`          | `false`      | HTTP server TLS enabled     |
| `--http-tls-cert-path`        | `OPEN-VE_HTTP_TLS_CERT_PATH`        | `""`         | HTTP server TLS cert path   |
| `--http-tls-key-path`         | `OPEN-VE_HTTP_TLS_KEY_PATH`         | `""`         | HTTP server TLS key path    |
| `--grpc-addr`                 | `OPEN-VE_GRPC_ADDR`                 | `:9000`      | gRPC server address         |
| `--grpc-tls-enabled`          | `OPEN-VE_GRPC_TLS_ENABLED`          | `false`      | gRPC server TLS enabled     |
| `--grpc-tls-cert-path`        | `OPEN-VE_GRPC_TLS_CERT_PATH`        | `""`         | gRPC server TLS cert path   |
| `--grpc-tls-key-path`         | `OPEN-VE_GRPC_TLS_KEY_PATH`         | `""`         | gRPC server TLS key path    |
| `--store-engine`              | `OPEN-VE_STORE_ENGINE`              | `redis`      | store engine (redis/memory) |
| `--store-redis-addr`          | `OPEN-VE_STORE_REDIS_ADDR`          | `redis:6379` | Redis address               |
| `--store-redis-password`      | `OPEN-VE_STORE_REDIS_PASSWORD`      | `""`         | Redis password              |
| `--store-redis-db`            | `OPEN-VE_STORE_REDIS_DB`            | `0`          | Redis DB                    |
| `--store-redis-pool-size`     | `OPEN-VE_STORE_REDIS_POOL_SIZE`     | `1000`       | Redis pool size             |
| `--log-level`                 | `OPEN-VE_LOG_LEVEL`                 | `info`       | Log level                   |

#### 2. Config File

You can also use a config file in YAML format.

Place the `config.yaml` in the same directory or `$HOME/.open-ve/config.yaml`.

```yaml
http:
  addr: ":8080"
  corsAllowedOrigins: ["*"]
  corsAllowedHeaders: ["*"]
  tls:
    enabled: false
    certPath: ""
    keyPath: ""
grpc:
  addr: ":9000"
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

#### 1. Build From Source

```bash
go build -o open-ve go/cmd/open-ve/main.go
./open-ve run
```

#### 2. Docker Compose

```bash
docker-compose up
```

## CEL

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

## Example (HTTP API)

### Register Validation Rules

Request:

```bash
curl --request POST \
  --url http://localhost:8080/v1/dsl \
  --header 'Content-Type: application/json' \
  --data '{
    "validations": [
      {
        "id": "item",
        "cels": [
          "price > 0", # price must be greater than 0
          "size(image) < 360" # image size must be less than 360 bytes
        ],
        "variables": [
          {
            "name": "price",
            "type": "int"
          },
          {
            "name": "image",
            "type": "bytes"
          }
        ]
      },
      {
        "id": "user",
        "cels": [
          "size(name) < 20" # name length must be less than 20
        ],
        "variables": [
          {
            "name": "name",
            "type": "string"
          }
        ]
		 }
    ]
  }'
```

Response:

```json
{}
```

### Get Current Validation Rules

Request:

```bash
curl --request GET \
  --url http://localhost:8080/v1/dsl \
  --header 'Content-Type: application/json'
```

Response:

```json
{
  "validations": [
    {
      "id": "item",
      "cels": ["price > 0", "size(image) < 360"],
      "variables": [
        {
          "name": "price",
          "type": "int"
        },
        {
          "name": "image",
          "type": "bytes"
        }
      ]
    },
    {
      "id": "user",
      "cels": ["size(name) < 20"],
      "variables": [
        {
          "name": "name",
          "type": "string"
        }
      ]
    }
  ]
}
```

### Validate

Request:

```bash
curl --request POST \
  --url 'http://localhost:8080/v1/check' \
  --header 'Content-Type: application/json' \
  --data '{
    "validations": [
      {
        "id": "item",
        "variables": {
          "price": -100,
          "image": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVR4nGO4unY2AAR4Ah51j5XwAAAAAElFTkSuQmCC" # send base64 encoded image
        }
      },
      {
        "id": "user",
        "variables": {
          "name": "longlonglonglongname"
        }
      }
    ]
  }'

```

Response:

```json
{
  "results": [
    {
      "id": "item",
      "isValid": false,
      "message": "failed validations: price > 0"
    },
    {
      "id": "user",
      "isValid": false,
      "message": "failed validations: size(name) < 20"
    }
  ]
}
```
