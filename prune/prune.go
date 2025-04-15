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
	for key := range cfg.CustomQueues.Queues {
		addPendingLastUnix := key + ":lastUnixCheck"
		userqueuedJobs, err := cronlisting.GetQueuedTasks(key)
		if err != nil {
			log.Println("Error on GetQueuedTasks for queue:", key, err)
			continue
		}
		start, err := redisClient.Get(context.Background(), addPendingLastUnix).Result()
		if err != nil {
			log.Println("Error in getting AddPendingLastUnix for queue:", key, err)
			continue
		}
		startUTC, err := strconv.Atoi(start)
		if err != nil {
			log.Println("Error in conerting AddPendingLastUnix value to int for queue:", key, err)
			continue
		}
		ctx := context.Background()
		startunixT := time.Unix(int64(startUTC), 0).UTC()
		redisClient.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
			for _, c := range *userqueuedJobs {
				cronUTC, err := c.GetUTC(startunixT)
				if err != nil {
					continue
				}
				// ready to prune
				if cronUTC < int64(startUTC) {
					pipeline.Pipeline().HDel(ctx, key, c.Name)
				}
			}
			return nil
		})
	}
}
