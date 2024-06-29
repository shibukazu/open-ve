package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
	"github.com/shibukazu/open-ve/go/pkg/validator"
)

func main() {
	var ctx = context.Background()

	redis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
		PoolSize: 1000,
	})

	dslReader := dsl.NewDSLReader(redis)

	data, err := os.ReadFile("dsl.yml")
	if err != nil {
		log.Fatal(err)
	}

	dslReader.Read(ctx, data)

	validator := validator.NewValidator(redis)
	res, err := validator.Validate("x-price", map[string]interface{}{
		"num": 100,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}
