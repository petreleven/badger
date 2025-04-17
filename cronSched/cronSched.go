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
	redisClient *redis.Client
	cfg         *config.Config
)

func AddPendingCronsStart() {
	redisClient = dbRedis.Get()
	cfg = config.Get()

	addPendingCrons()
	go func() {
		for {
			time.Sleep(30 * time.Second)
			addPendingCrons()
		}
	}()
}

// Assuming a cluster of workers , try to take charge of adding cron jobs to queue
func addPendingCrons() error {
	ctx := context.Background()
	for key := range cfg.CustomQueues.Queues {
		var (
			addToPendingLock    = key + ":workerInCharge"
			addToPendingLockTTL = 90 * time.Second
			addPendingLastUnix  = key + ":lastUnixCheck"
		)
		var (
			setCmd *redis.BoolCmd
			getCmd *redis.StringCmd
		)

		value := ""
		// try to take charge
		// first make sure the key doesnt exist
		_, err := redisClient.Get(ctx, addToPendingLock).Result()
		if err != redis.Nil {
			return err
		}
		_, err = redisClient.Pipelined(context.Background(), func(pipeline redis.Pipeliner) error {
			setCmd = pipeline.SetNX(ctx, addToPendingLock, cfg.WorkerID, addToPendingLockTTL)
			getCmd = pipeline.Get(ctx, addToPendingLock)
			return nil
		})
		if err != nil {
			log.Printf("Pipeline failed: %v\n", err)
			return err
		}

		_, err = setCmd.Result()
		if err != nil {
			log.Printf("ERROR TRYING TO TAKE CHARGE OF addToPendingLock:%v\n", err)
			return err
		}
		value, err = getCmd.Result()
		if err != nil {
			log.Printf("ERROR TRYING TO GET addToPendingLock%v\n", err)
			return err
		}

		// Check if we took role of cron scheduling
		if value != cfg.WorkerID {
			// log.Println("Cant take role of cronmaster")
			return nil
		}
		// Get currentTime
		end, err := redisClient.Time(ctx).Result()
		if err != nil {
			end = time.Now()
		}
		end = ZeroOutSecondAndNanoSecond(end)
		// Get Last Check time
		startStr, err := redisClient.Get(ctx, addPendingLastUnix).Result()
		var startInt int64
		if err == redis.Nil {
			startInt = end.Unix()
		} else if err != nil {
			log.Printf("Failed to get addPendingLastUnix: %v\n", err)
			return err
		} else {
			startInt, err = strconv.ParseInt(startStr, 10, 64)
			if err != nil {
				log.Printf("failed to parse addPendingLastUnix value: %v\n", err)
				return err
			}
		}
		start := time.Unix(startInt, 0)
		start = ZeroOutSecondAndNanoSecond(start)
		if end.Sub(start) > 24*time.Hour || start.After(end) {
			start = end
		}

		userlisting, err := cronlisting.GetQueuedTasks(key)
		if userlisting == nil || err != nil {
			return err
		}
		_, err = redisClient.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
			for _, cron := range *userlisting {
				t := start
				for t.Compare(end) <= 0 {
					var status bool
					status, err = IsCronReady(&cron, t)
					if err != nil {
						log.Println("Error checking if cron:", cron.Name, " ISready ", err)
						break
					}

					if status {
						jb := cron.Json()
						pipeline.LPush(ctx, "badger:pending:"+key, cron.Name+":"+string(jb))
						break
					}
					t = t.Add(1 * time.Minute)
				}

			}
			return nil
		})

		nextEnd := end.Add(1 * time.Minute)
		_, err = redisClient.Set(ctx, addPendingLastUnix, strconv.FormatInt(nextEnd.Unix(), 10), -1).Result()
		if err != nil {
			log.Println("ERROR SETTING addpendingLastUnix ", err)
		}
		// redisClient.Del(ctx, addToPendingLock).Result()
		return nil
	}
	return nil
}

func ZeroOutSecondAndNanoSecond(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

func IsCronReady(c *cronlisting.Cron, t time.Time) (bool, error) {
	cutc, err := c.GetUTC(t)
	if err != nil {
		return false, err
	}
	if cutc == t.UTC().Unix() {
		return true, nil
	}
	return false, nil
}
