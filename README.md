# Open-VE: Centralized and Consistent Data Validation Engine

![GitHub Release](https://img.shields.io/github/v/release/shibukazu/open-ve)
![GitHub License](https://img.shields.io/github/license/shibukazu/open-ve)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/shibukazu/open-ve)
![GitHub Repo stars](https://img.shields.io/github/stars/shibukazu/open-ve)

A powerful solution that **simplifies the management of validation rules**, ensuring consistent validation across all layers, including frontend, BFF, and microservices, through a single, simple API.

Open-VE offers an **HTTP API** and a **gRPC API**. We currently provide a [Go](https://github.com/shibukazu/open-ve-go-sdk) and [TypeScript](https://github.com/shibukazu/open-ve-typescript-sdk) SDK.

## Features

### 📕 Centralized Validation Logic Management

Manage validation rules in one place using Common Expression Language (CEL), ensuring language-agnostic consistency of validation across your system.

### 🔌 API-Based Validation Management and Query

Register, update, retrieve, and query validation rules through a simple and consistent API, enabling seamless validation checks and eliminating the need for custom logic at various application layers.

### 🏭 Schema Auto Generation

Generate Open-VE Schame from OpenAPI and Protobuf definitions for ease of integration.

### 🧪 Schema Unit Testing

Validate Open-VE schema correctness with built-in unit testing capabilities.

### 🌐 Distributed Validation Logic Management for Microservices

Supports master-slave architecture for distributed validation management. Improve scalability and compatibility with microservice environments.

## Getting Started

### brew

```bash
brew install shibukazu/tap/open-ve
open-ve run
```

### Build From Source

```bash
go build -o open-ve go/cmd/open-ve/main.go
./open-ve run
```

### Docker Compose

```bash
docker-compose up
```

## Reference

### Example

- [Example of Master Slave Architecture](docs/Master-Slave-Example.md)
- [Example of Monolithic Architecture](docs/Monolithic-Example.md)

### Documents

- [Config](docs/Config.md)
- [TLS](docs/TLS.md)
- [Performance](docs/Performance.md)

## Limitation

### CEL

We only support the basic types of CEL currently.
| Type | Support | Future Support |
| ------------- | ------- | -------------- |
| `int` | ✅ | |
| `uint` | ✅ | |
| `double` | ✅ | |
| `bool` | ✅ | |
| `string` | ✅ | |
| `bytes` | ✅ | |
| `list` | | ✅ |
| `map` | | ✅ |
| `null_type` | | ❓ |
| message names | | ❓ |
| `type` | | ❓ |
