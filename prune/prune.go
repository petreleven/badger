package prune

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"worker/config"
	"worker/cronlisting"
	"worker/dbRedis"
)

func PruneStart() {
	prune()
	go func() {
		for {
			time.Sleep(30 * time.Second)
			prune()
		}
	}()
}

func prune() {
	redisClient := dbRedis.Get()
	cfg := config.Get()
	userqueuedJobs, err := cronlisting.GetUserQueuedTasks()
	if err != nil {
		log.Println("Error on GetUserQueuedTasks ", err)
		return
	}
	start, err := redisClient.Get(context.Background(), cfg.AddPendingLastUnix).Result()
	if err != nil {
		log.Println("Error in getting AddPendingLastUnix", err)
		return
	}
	startUTC, err := strconv.Atoi(start)
	if err != nil {
		log.Println("Error in conerting AddPendingLastUnix value to int", err)
		return
	}
	ctx := context.Background()
	redisClient.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
		for _, c := range *userqueuedJobs {
			cronUTC, err := c.GetUTC()
			if err != nil {
				continue
			}
			// ready to prune
			if cronUTC < int64(startUTC) {
				pipeline.Pipeline().HDel(ctx, cfg.UserQueue, c.Name)
			}
		}
		return nil
	})
}
