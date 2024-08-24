# Master-Slave Example

This example demonstrates how to run the Open-VE as master-slave architecture.

Note: In this example, the master node is hosted at `localhost:8081`, and the slave node is hosted at `localhost:8082`.

## Run

### Master Node

```bash
open-ve run --mode=master --http-port=8081 --grpc-port=9001
```

### Slave Node

```bash
open-ve run --mode=slave --slave-id=slave-node-id --slave-master-http-addr=http://localhost:8081 --slave-slave-http-addr=http://localhost:8082 --http-port=8082 --grpc-port=9002
```

## Scinario

### 1. Register a Set of Validation Rules to Master Node

```bash
curl --request POST \
  --url http://localhost:8081/v1/dsl \
  --header 'Content-Type: application/json' \
  --data '{
	"validations": [
		{
			"id": "user",
			"cels": [
				"size(name) < 20"
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

### 2. Register a Set of Validation Rules to Slave Node

```bash
curl --request POST \
  --url http://localhost:8082/v1/dsl \
  --header 'Content-Type: application/json' \
  --data '{
	"validations": [
		{
			"id": "item",
			"cels": [
				"price > 0",
				"size(image) < 360"
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
		}
	]
}'
```

### 3. Check the Current Validation Rules

```bash
curl --request GET \
  --url http://localhost:8081/v1/dsl \
  --header 'Content-Type: application/json'
```

```json
{
  "validations": [
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

```bash
curl --request GET \
  --url http://localhost:8082/v1/dsl \
  --header 'Content-Type: application/json'
```

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
    }
  ]
}
```

### 4. Request Validation Check to Master Node

Although only part of the validation rules are registered with the master node, you can request validation for all rules, including those on the slave nodes.

```bash
curl --request POST \
  --url http://localhost:8081/v1/check \
  --header 'Content-Type: application/json' \
  --data '{
	"validations": [
		{
			"id": "item",
			"variables": {
				"price": -100,
				"image": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADElEQVR4nGO4unY2AAR4Ah51j5XwAAAAAElFTkSuQmCC"
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

```json
{
  "results": [
    {
      "id": "user",
      "isValid": false,
      "message": "failed validations: size(name) < 20"
    },
    {
      "id": "item",
      "isValid": false,
      "message": "failed validations: price > 0"
    }
  ]
}
```
