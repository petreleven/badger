package prune

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"worker/config"
	"worker/cronlisting"
	"worker/dbRedis"
	hb "worker/heartbeat"
)

var (
	redisClient *redis.Client
	cfg         *config.Config
)

func PruneStart() {
	redisClient = dbRedis.Get()
	cfg = config.Get()
	pruneCustomQueues()
	pruneZombieWorkers()
	go func() {
		for {
			time.Sleep(30 * time.Second)
			pruneCustomQueues()
		}
	}()
	go func() {
		for {
			time.Sleep(30 * time.Second)
			pruneZombieWorkers()
		}
	}()
}

func pruneCustomQueues() {
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

func pruneZombieWorkers() {
	ctx := context.Background()

	result, err := redisClient.HGetAll(ctx, cfg.ClusterName).Result()
	if err != nil && err != redis.Nil {
		log.Println("Unable to get the workers in cluster for pruning ", err)
		return
	}
	t := time.Now().Add(-5 * time.Minute).Unix()
	workersToPrune := []string{}
	for key, value := range result {
		hbMetaData := hb.WorkerMetaData{}
		err = json.Unmarshal([]byte(value), &hbMetaData)
		if err != nil {
			log.Println("Unable to decode worker metadata ", err)
			continue
		}
		if hbMetaData.Timestamp <= t {
			workersToPrune = append(workersToPrune, key)
		}
	}
	_, err = redisClient.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
		for _, value := range workersToPrune {
			redisClient.HDel(ctx, cfg.ClusterName, value)
		}
		return nil
	})
	if err != nil {
		log.Println("Unable to do piplene in fn pruneZombieWorkers  ", err)
	}
}
