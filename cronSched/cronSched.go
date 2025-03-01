package cronSched

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

var (
	redisClient *redis.Client  = dbRedis.Get()
	cfg         *config.Config = config.Get()
)

func AddPendingCronsStart() {
	addPendingCrons()
	go func() {
		for {
			time.Sleep(30 * time.Second)
			addPendingCrons()
		}
	}()
}

func addPendingCrons() error {
	// Assuming a cluster of workers , try to take charge of adding cron jobs to queue
	ctx := context.Background()
	var (
		addpendingMasterKey = cfg.AddPendingMasterKey
		cronMasterTTL       = 90 * time.Second
		addPendingLastUnix  = cfg.AddPendingLastUnix
	)
	value := ""
	var (
		setCmd *redis.BoolCmd
		getCmd *redis.StringCmd
	)
	_, err := redisClient.Pipelined(context.Background(), func(pipeline redis.Pipeliner) error {
		setCmd = pipeline.SetNX(ctx, addpendingMasterKey, cfg.WorkerProcName, cronMasterTTL)
		getCmd = pipeline.Get(ctx, addpendingMasterKey)
		return nil
	})
	if err != nil {
		log.Printf("Pipeline failed: %v\n", err)
		return err
	}

	_, err = setCmd.Result()
	if err != nil {
		log.Printf("ERROR TRYING TO TAKE CHARGE OF addpendingMasterKey:%v\n", err)
		return err
	}
	value, err = getCmd.Result()
	if err != nil {
		log.Printf("ERROR TRYING TO GET addpendingMasterKey%v\n", err)
		return err
	}

	// Check if we took role of cron scheduling
	if value != cfg.WorkerProcName {
		log.Println("Cant take role of cronmaster")
		return nil
	}
	// TODO push jobs to pending queue
	end, err := redisClient.Time(ctx).Result()
	if err != nil {
		end = time.Now()
	}
	end = ZeroOutSecondAndNanoSecon(end)
	startStr, err := redisClient.Get(ctx, addPendingLastUnix).Result()
	var startInt int64
	if err == redis.Nil {
		startInt = end.Unix()
	} else if err != nil {
		log.Printf("failed to get addPendingLastUnix: %v\n", err)
		return err
	} else {
		startInt, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			log.Printf("failed to parse addPendingLastUnix value: %v\n", err)
			return err
		}
	}
	start := time.Unix(startInt, 0)
	start = ZeroOutSecondAndNanoSecon(start)
	if end.Sub(start) > 24*time.Hour || start.After(end) {
		start = end
	}

	userlisting, err := cronlisting.GetUserQueuedTasks()
	if userlisting == nil {
		return err
	}
	_, err = redisClient.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
		// TODO- add jobs to pending queue if their schedtime minute is between start -> end
		for _, cron := range *userlisting {
			data := cron.Json()
			pipeline.LPush(ctx, cron.Queue, string(data))
		}
		return nil
	})

	nextEnd := end.Add(time.Minute)
	_, err = redisClient.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
		pipeline.Set(ctx, addPendingLastUnix, strconv.FormatInt(nextEnd.Unix(), 10), 0)
		return nil
	})
	if err != nil {
		log.Println("ERROR SETTING addpendingLastUnix ", err)
	}
	return nil
}

func ZeroOutSecondAndNanoSecon(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}
