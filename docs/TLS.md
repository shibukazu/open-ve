# TLS

This document describes how to set up TLS connection in Open-VE.

> [!NOTE]
> This document uses docker-compose to run Open-VE. So you can connct slave node by using `localhost:8082` from outside of the container, and `slave-node:8082` from inside of the container.

## Generate Certificates

Run following commands to generate certificates.

```bash
mkdir -p local/ssl/generated

# CAのプライベートキーを生成
openssl genrsa -out local/ssl/generated/ca.key 2048

# Master Nodeのプライベートキーを生成
openssl genrsa -out local/ssl/generated/master.key 2048

# Slave Nodeのプライベートキーを生成
openssl genrsa -out local/ssl/generated/slave.key 2048

# CAの自己署名証明書を生成
openssl req -x509 -new -nodes -key local/ssl/generated/ca.key -sha256 -days 365 -out local/ssl/generated/ca.crt -subj "/C=US/ST=State/L=City/O=Organization/OU=Unit/CN=my-ca"

# Master NodeのCSRを生成
openssl req -new -key local/ssl/generated/master.key -out local/ssl/generated/master.csr -config local/ssl/cnf/master.cnf

# Slave NodeのCSRを生成
openssl req -new -key local/ssl/generated/slave.key -out local/ssl/generated/slave.csr -config local/ssl/cnf/slave.cnf

# CAによってMaster Nodeの証明書に署名
openssl x509 -req -in local/ssl/generated/master.csr -CA local/ssl/generated/ca.crt -CAkey local/ssl/generated/ca.key -CAcreateserial -out local/ssl/generated/master.crt -days 365 -sha256 -extfile local/ssl/cnf/master.cnf -extensions v3_req

# CAによってSlave Nodeの証明書に署名
openssl x509 -req -in local/ssl/generated/slave.csr -CA local/ssl/generated/ca.crt -CAkey local/ssl/generated/ca.key -CAcreateserial -out local/ssl/generated/slave.crt -days 365 -sha256 -extfile local/ssl/cnf/slave.cnf -extensions v3_req
```

## Config

Second, update docker-compose.yml to enable TLS.

```yaml
services:
  redis:
    image: redis:latest
    container_name: redis
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - default
    restart: unless-stopped
  master-node:
    build:
      context: .
    container_name: master-node
    ports:
      - "8081:8080"
      - "9001:9000"
    volumes:
      - ./local:/local
    networks:
      - default
    depends_on:
      - redis
    environment:
      - OPEN-VE_MODE=master
      - OPEN-VE_HTTP_PORT=
      - OPEN-VE_HTTP_CORS_ALLOWED_ORIGINS=
      - OPEN-VE_HTTP_CORS_ALLOWED_HEADERS=
      - OPEN-VE_HTTP_TLS_ENABLED=true
      - OPEN-VE_HTTP_TLS_CERT_PATH=/local/ssl/generated/master.crt
      - OPEN-VE_HTTP_TLS_KEY_PATH=/local/ssl/generated/master.key
      - OPEN-VE_GRPC_PORT=
      - OPEN-VE_GRPC_TLS_ENABLED=true
      - OPEN-VE_GRPC_TLS_CERT_PATH=/local/ssl/generated/master.crt
      - OPEN-VE_GRPC_TLS_KEY_PATH=/local/ssl/generated/master.key
      - OPEN-VE_STORE_ENGINE=redis
      - OPEN-VE_STORE_REDIS_ADDR=
      - OPEN-VE_STORE_REDIS_PASSWORD=
      - OPEN-VE_STORE_REDIS_DB=
      - OPEN-VE_STORE_REDIS_POOL_SIZE=
      - OPEN-VE_LOG_LEVEL=
  slave-node:
    build:
      context: .
    container_name: slave-node
    ports:
      - "8082:8080"
      - "9002:9000"
    volumes:
      - ./local:/local
    networks:
      - default
    depends_on:
      - redis
      - master-node
    environment:
      - OPEN-VE_MODE=slave
      - OPEN-VE_SLAVE_ID=slave-node
      - OPEN-VE_SLAVE_SLAVE_HTTP_ADDR=https://slave-node:8080
      - OPEN-VE_SLAVE_MASTER_HTTP_ADDR=https://master-node:8080
      - OPEN-VE_HTTP_PORT=
      - OPEN-VE_HTTP_CORS_ALLOWED_ORIGINS=
      - OPEN-VE_HTTP_CORS_ALLOWED_HEADERS=
      - OPEN-VE_HTTP_TLS_ENABLED=true
      - OPEN-VE_HTTP_TLS_CERT_PATH=/local/ssl/generated/slave.crt
      - OPEN-VE_HTTP_TLS_KEY_PATH=/local/ssl/generated/slave.key
      - OPEN-VE_GRPC_PORT=
      - OPEN-VE_GRPC_TLS_ENABLED=true
      - OPEN-VE_GRPC_TLS_CERT_PATH=/local/ssl/generated/slave.crt
      - OPEN-VE_GRPC_TLS_KEY_PATH=/local/ssl/generated/slave.key
      - OPEN-VE_STORE_ENGINE=redis
      - OPEN-VE_STORE_REDIS_ADDR=
      - OPEN-VE_STORE_REDIS_PASSWORD=
      - OPEN-VE_STORE_REDIS_DB=
      - OPEN-VE_STORE_REDIS_POOL_SIZE=
      - OPEN-VE_LOG_LEVEL=
  apidocs:
    image: swaggerapi/swagger-ui
    container_name: apidocs
    environment:
      - SWAGGER_JSON=/openapi/openapi.swagger.json
      - BASE_URL=/docs
    volumes:
      - ./openapi:/openapi
    restart: unless-stopped
    ports:
      - "18080:8080"

volumes:
  redis-data:

networks:
  default:
    driver: bridge
```

## Patch

Finally, update `go/pkg/server/gateway.go` and `go/pkg/slave/registrar.go` to allow self-signed certificates.

```go
// go/pkg/server/gateway.go
if slaveNode.TLSEnabled {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true, // here
        },
    }
    client = &http.Client{Transport: transport}
} else {
    client = &http.Client{}
}
```

```go
// go/pkg/slave/registrar.go
if masterTLSEnabled {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true, // here
        },
    }
    client = &http.Client{Transport: transport}
} else {
    client = &http.Client{}
}
```
