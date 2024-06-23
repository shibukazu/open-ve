package main

import (
	"context"
	"log"
	"os"

	"github.com/go-redis/redis"
	"github.com/shibukazu/open-ve/go/pkg/dsl"
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
}
