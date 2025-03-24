package cronSched

import (
	"context"
	"log"
	"strconv"
	"strings"
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

	// try to take charge
	// first make sure the key doesnt exist
	_, err := redisClient.Get(ctx, addpendingMasterKey).Result()
	if err != redis.Nil {
		return err
	}
	_, err = redisClient.Pipelined(context.Background(), func(pipeline redis.Pipeliner) error {
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
		// log.Println("Cant take role of cronmaster")
		return nil
	}
	// TODO push jobs to pending queue
	end, err := redisClient.Time(ctx).Result()
	if err != nil {
		end = time.Now()
	}
	end = ZeroOutSecondAndNanoSecond(end)
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

	userlisting, err := cronlisting.GetUserQueuedTasks()
	if userlisting == nil || err != nil {
		return err
	}
	_, err = redisClient.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
		for _, cron := range *userlisting {
			t := start
			for t.Compare(end) <= 0 {
				status, err := IsCronReady(&cron, t)
				if err != nil {
					log.Println("Error checking if cron:", cron.Name, " ISready ", err)
					break
				}
				if status {
					jb := cron.Json()
					pipeline.LPush(ctx, cron.Queue, cron.Name+":"+string(jb))
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
	// redisClient.Del(ctx, addpendingMasterKey).Result()
	return nil
}

func ZeroOutSecondAndNanoSecond(t time.Time) time.Time {
	return t.Truncate(time.Minute)
}

func isInRange(value string, t int) (bool, error) {
	if value == "*" {
		return true, nil
	}

	if strings.HasPrefix(value, "*/") {
		if len(value) > 2 {
			divisor, err := strconv.Atoi(value[2:])
			if err != nil {
				return false, err
			}
			if t%divisor == 0 {
				return true, nil
			}
		}
		return false, nil
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return false, err
	}
	if v == t {
		return true, nil
	}

	return false, nil
}

func IsCronReady(c *cronlisting.Cron, t time.Time) (bool, error) {
	ready, err := isInRange(c.DayWeek, int(t.Weekday()))
	if err != nil || !ready {
		return false, err
	}
	ready, err = isInRange(c.Month, int(t.Month()))
	if err != nil || !ready {
		return false, err
	}
	ready, err = isInRange(c.Day, t.Day())
	if err != nil || !ready {
		return false, err
	}
	ready, err = isInRange(c.Hour, t.Hour())
	if err != nil || !ready {
		return false, err
	}
	ready, err = isInRange(c.Minute, t.Minute())
	if err != nil || !ready {
		return false, err
	}
	return true, nil
}
