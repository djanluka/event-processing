package redis

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	once        sync.Once
)

func GetRedisClient() *redis.Client {
	once.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     "redis:6379", // Redis server address
			Password: "",           // No password set
			DB:       0,            // Use default DB
		})

		// Ping the Redis server to check if the connection is successful
		_, err := redisClient.Ping(context.Background()).Result()
		if err != nil {
			panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
		}
	})

	return redisClient
}

func Close() {
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}
	log.Println("Redis connection closed")
}
