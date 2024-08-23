# Monolithic Example

This example demonstrates how to run the Open-VE as monolithic architecture.

## Run

```bash
open-ve run
```

## Scenario

### 1. Register Validation Rules

```bash
curl --request POST \
  --url http://localhost:8080/v1/dsl \
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

### 2. Check the Current Validation Rules

```bash
curl --request GET \
  --url http://localhost:8080/v1/dsl \
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

### 3. Request Validation Check

```bash
curl --request POST \
  --url http://localhost:8080/v1/check \
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
