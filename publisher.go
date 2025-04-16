package main

import (
	"context"
	"flag"
	"log"
	"time"

	listing "worker/cronlisting"

	"github.com/redis/go-redis/v9"
)

var (
	cmd      *int    = flag.Int("t", 0, "get 0, set 1")
	RedisUrl *string = flag.String("u", "redis://localhost:6739", "redis url")
)

func main() {
	flag.Parse()
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
		PoolSize:         2,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     5 * time.Second,
		PoolTimeout:      1 * time.Second,
	}

	if err != nil {
		log.Println("Cant parse url")
		return
	}
	redisClient := redis.NewClient(options)
	/*
	 */
	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalln("Failed to connect to redis: ", err)
	}
	log.Println("Connected to Redis")

	ctx := context.Background()
	c1 := listing.Cron{
		Name:    "start",
		Minute:  "20",
		Hour:    "20",
		Day:     "01",
		Month:   "12",
		DayWeek: "*",
		Job:     "run bash",
		Queue:   "userqueue",
	}

	if *cmd == 1 {
		_, err := redisClient.HSet(ctx, "mycustomqueue",
			c1.Name, c1.Encode()).Result()
		if err != nil {
			log.Println("Unable to save to UserQueue: ", err)
		}
		log.Println("sent")
	}

	r, err := redisClient.HGetAll(ctx, "mycustomqueue").Result()
	log.Println(r)
}
