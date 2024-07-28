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

You can overwrite the default configuration. Create a `config.yaml` file in the root directory of the project.

```yaml
http:
  addr: "localhost:8080"
  cordAllowedOrigins: "*"
  corsAllowedMethods: "*"
grpc:
  addr: "localhost:9000"
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  poolSize: 10
log:
  level: "info"
```

### Run

```bash
docker compose up
```

## CEL

We use [CEL](https://github.com/google/cel-spec/blob/master/doc/langdef.md) as the expression language for validation rules.

Supported types:

| Type          | Support | Future Support |
| ------------- | ------- | -------------- |
| `int`         | ✅      |                |
| `uint`        | ✅      |                |
| `double`      | ✅      |                |
| `bool`        | ✅      | 　             |
| `string`      | ✅      | 　             |
| `bytes`       | ✅      | 　             |
| `list`        |         | 　　 ✅        |
| `map`         |         | 　 ✅          |
| `null_type`   |         | 　 ❓          |
| message names |         | 　 ❓          |
| `type`        |         | 　 ❓          |

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
        "id": "price",
        "cel": "number % 3 == 0 || number < 5",
        "variables": [
          {
            "name": "number",
            "type": "int"
          }
        ]
      },
    ]
  }'
```

Response:

```bash
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

```bash
{
	"validations": [
		{
			"id": "price",
			"cel": "number % 3 == 0 || number < 5",
			"variables": [
				{
					"name": "number",
					"type": "int"
				}
			]
		},
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
    "id": "price",
    "variables": {
      "number": 0
    }
  }'
```

Response:

```bash
{"isValid":false, "message":""}
```
