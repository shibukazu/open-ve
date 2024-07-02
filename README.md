# open-ve

Centralized and Consistent Data Validation Engine

## Note

This project is still under development and not ready for production use.

We only support limited CEL expression and gRPC API.

## Setup

### Config

If you want to specify the configuration below, you can create a `config.yaml` file in the root directory of the project.
If you don't specify the configuration, the default values will be used.

```yaml
server:
  grpc:
    port: 9000
  rest:
    port: 8080
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  poolSize: 10
```

### Redis

```bash
docker compose up -d
```

### Server

```bash
go run cmd/main.go
```

## Example

### Register Validation Rules

You may need to install `grpcurl` and `yq` before running the command below.

Save DSL below to a file named `local/dsl.yml`.

```yaml
validations:
  - id: "price"
    cel: "number % 3 == 0 || number < 5"
    variables:
      - name: "number"
        type: "int"
```

```bash
yq eval -o=json local/dsl.yml | grpcurl -plaintext -d @ localhost:9000 dsl.v1.DSLService/Register
```

### Read Validation Rules

```bash
grpcurl -plaintext -d '{}' localhost:9000 dsl.v1.DSLService/Read
```

### Validate Data

```bash
grpcurl -plaintext -d '{
  "id": "price",
  "variables": {
    "number": {
      "@type": "type.googleapis.com/google.protobuf.Int32Value",
      "value": 3
    }
  }
}' localhost:9000 validate.v1.ValidateService/Check
```
