package dbRedis

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func Get() *redis.Client {
	if redisClient == nil {
		redisClient = connectAndSave()
	}
	return redisClient
}

func connectAndSave() *redis.Client {
	redisClient = redis.NewClient(&redis.Options{
		Addr:             "localhost:6380",
		Password:         "",
		DB:               1,
		DisableIndentity: true,
		MaxRetries:       10,
		PoolSize:         2,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     5 * time.Second,
		PoolTimeout:      1 * time.Second,
	})
	err := redisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalln("Failed to connect to redis: ", err)
	}
	log.Println("Connected to Redis")
	return redisClient
}
