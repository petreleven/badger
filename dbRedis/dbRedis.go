package dbRedis

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	RedisUrl    *string
	lock        sync.RWMutex
)

func Get() *redis.Client {
	lock.Lock()
	defer lock.Unlock()
	if redisClient == nil {
		redisClient = connectAndSave()
	}
	return redisClient
}

func connectAndSave() *redis.Client {
	op, err := redis.ParseURL(*RedisUrl)
	if err != nil {
		log.Panicln("ERROR parsing redis url")
	}
	options := &redis.Options{
		Addr:             op.Addr,
		Password:         op.Password,
		DB:               0,
		DisableIndentity: true,
		MaxRetries:       10,
		PoolSize:         20,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     5 * time.Second,
		PoolTimeout:      60 * time.Second,
	}

	if err != nil {
		log.Println("Cant parse url")
		return nil
	}
	redisClient = redis.NewClient(options)
	/*
	 */
	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Println("Are you sure redis at ", *RedisUrl, "is Running")
		log.Fatalln("Failed to connect to redis: ", err)
	}
	log.Println("Connected to Redis")
	return redisClient
}
