package main

import (
	"context"
	"log"

	"worker/config"
	"worker/dbRedis"

	"github.com/redis/go-redis/v9"
	listing "worker/cronlisting"
)

var (
	redisClient *redis.Client  = dbRedis.Get()
	cfg2        *config.Config = config.Get()
)

func main2() {
	c2 := listing.Cron{
		Name:    "task1",
		Minute:  "*/5",
		Hour:    "*",
		Day:     "*",
		Month:   "*",
		DayWeek: "*",
	}

	c3 := listing.Cron{
		Name:    "task2",
		Minute:  "46",
		Hour:    "*",
		Day:     "*",
		Month:   "*",
		DayWeek: "*",
	}

	c4 := listing.Cron{
		Name:    "task3",
		Minute:  "47",
		Hour:    "13",
		Day:     "*",
		Month:   "*",
		DayWeek: "*",
	}

	c5 := listing.Cron{
		Name:    "task4",
		Minute:  "49",
		Hour:    "*/13",
		Day:     "*",
		Month:   "*",
		DayWeek: "1",
	}

	// Add all jobs to the queue
	jobs := []*listing.Cron{&c2, &c3, &c4, &c5}
	for _, job := range jobs {
		AddToUserQueue(job)
	}
}

func AddToUserQueue(c *listing.Cron) {
	ctx := context.Background()
	_, err := redisClient.HSet(ctx, cfg2.UserQueue,
		c.Name+":string", c.Encode()).Result()
	if err != nil {
		log.Println("Unable to save to UserQueue: ", err)
	}
}
