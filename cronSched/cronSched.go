package cronSched

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"worker/config"
	"worker/cronlisting"
	"worker/dbRedis"
)

var (
	redisClient *redis.Client
	cfg         *config.Config
)

type DoneLogWriter struct {
	QueueKey string
	JobId    string
}

func (w *DoneLogWriter) Write(p []byte) (n int, err error) {
	redisClient := dbRedis.Get()
	ctx := context.Background()

	existingLog, err := redisClient.HGet(ctx, "badger:joblog", w.JobId).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	if existingLog != "" {
		existingLog += "\n"
	}
	existingLog += string(p)

	if err := redisClient.HSet(ctx, "badger:joblog", w.JobId, existingLog).Err(); err != nil {
		return 0, err
	}
	return len(p), nil
}

func AddPendingCronsStart() {
	redisClient = dbRedis.Get()
	cfg = config.Get()

	go func() {
		for {
			addPendingCrons()
			time.Sleep(10 * time.Second)
		}
	}()
	startWorkers()
}

// Assuming a cluster of workers , try to take charge of adding cron jobs to queue
func addPendingCrons() {
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
			continue
		}
		_, err = redisClient.Pipelined(context.Background(), func(pipeline redis.Pipeliner) error {
			setCmd = pipeline.SetNX(ctx, addToPendingLock, cfg.WorkerID, addToPendingLockTTL)
			getCmd = pipeline.Get(ctx, addToPendingLock)
			return nil
		})
		if err != nil {
			log.Printf("Pipeline failed: %v\n", err)
			continue
		}

		_, err = setCmd.Result()
		if err != nil {
			log.Printf("ERROR TRYING TO TAKE CHARGE OF addToPendingLock:%v\n", err)
			continue
		}
		value, err = getCmd.Result()
		if err != nil {
			log.Printf("ERROR TRYING TO GET addToPendingLock%v\n", err)
			continue
		}

		// Check if we took role of cron scheduling
		if value != cfg.WorkerID {
			// log.Println("Cant take role of cronmaster")
			continue
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
			continue
		} else {
			startInt, err = strconv.ParseInt(startStr, 10, 64)
			if err != nil {
				log.Printf("failed to parse addPendingLastUnix value: %v\n", err)
				continue
			}
		}
		start := time.Unix(startInt, 0)
		start = ZeroOutSecondAndNanoSecond(start)
		if end.Sub(start) > 24*time.Hour || start.After(end) {
			start = end
		}
		userlisting, err := cronlisting.GetQueuedTasks(key)
		if userlisting == nil || err != nil {
			continue
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
						pipeline.LPush(ctx, "badger:pending:"+key, string(jb))
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
	}
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

func worker(queueKey string, value config.CustomQueue, workerID string) {
	ctx := context.Background()
	cfg := config.Get()

	pendingQueue := "badger:pending:" + queueKey
	runningHash := "badger:running:" + queueKey
	delayedQueue := "badger:delayed:" + queueKey
	failedQueue := "badger:failed:" + queueKey
	doneQueue := "badger:done:" + queueKey

	// Try to get a job from the pending queue
	job, err := redisClient.BRPop(ctx, 10*time.Second, pendingQueue).Result()
	if err != nil {
		if err != redis.Nil {
			log.Printf("Worker %s: Unable to pop from %s: %v", workerID, pendingQueue, err)
		}
		return
	}

	// job[0] is the queue name, job[1] is the actual job
	jobCmd := job[1]
	jobID := uuid.New().String()

	// Use a transaction to atomically add the job to the running hash
	var setcmd *redis.IntCmd
	workerKey := fmt.Sprintf("%s:%s", workerID, jobID)
	setcmd = redisClient.HSet(ctx, runningHash, workerKey, jobCmd)
	_, err = setcmd.Result()
	if err != nil {
		log.Printf("Worker %s: Failed to add job to running hash: %v", workerID, err)
		// Put job back to pending
		redisClient.LPush(ctx, pendingQueue, jobCmd)
		return
	}

	cmd := exec.Command("bash", "-c", jobCmd)
	doneWriter := &DoneLogWriter{QueueKey: queueKey, JobId: jobID}
	log.Println(cfg.CustomQueues.Queues[queueKey].DoneLog)
	if cfg.CustomQueues.Queues[queueKey].DoneLog {
		cmd.Stdout = doneWriter
		cmd.Stderr = doneWriter
	} else {
		cmd.Stdout = log.Writer()
		cmd.Stderr = log.Writer()
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	elapseDuration := time.Duration(value.Timeout) * time.Second

	err = cmd.Start()
	if err != nil {
		log.Printf("Worker %s: Failed to start command: %v", workerID, err)
		// Remove from running hash and move to failed
		redisClient.HDel(ctx, runningHash, workerKey)
		redisClient.LPush(ctx, failedQueue, jobCmd)
		return
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err = <-done:
		// Remove from running hash regardless of outcome
		redisClient.HDel(ctx, runningHash, workerKey)

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitcode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
				if exitcode == 75 { // unix temp failure
					log.Printf("Worker %s: Job temporarily failed, moving to delayed", workerID)
					redisClient.LPush(ctx, delayedQueue, jobID+":"+jobCmd)
				} else {
					log.Printf("Worker %s: Job failed with exit code: %d", workerID, exitcode)
					redisClient.LPush(ctx, failedQueue, jobID+":"+jobCmd)
				}
			} else {
				log.Printf("Worker %s: Job failed with error: %v", workerID, err)
				redisClient.LPush(ctx, failedQueue, jobID+":"+jobCmd)
			}
		} else {
			log.Printf("Worker %s: Job completed successfully", workerID)
			redisClient.LPush(ctx, doneQueue, jobID+":"+jobCmd)
		}

	case <-time.After(elapseDuration):
		log.Printf("Worker %s: Job timed out after %v seconds", workerID, value.Timeout)

		// Kill process group
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			log.Printf("Worker %s: Failed to kill process: %v", workerID, err)
		}

		// Remove from running hash and move to failed
		redisClient.HDel(ctx, runningHash, workerKey)
		redisClient.LPush(ctx, failedQueue, jobID+":"+jobCmd)
	}
}

func startWorkers() {
	// Generate a unique process ID for this instance
	processID := uuid.New().String()

	// Start workers for each queue
	for queueKey, queueConfig := range cfg.CustomQueues.Queues {
		concurrency := queueConfig.Concurrency
		if concurrency <= 0 {
			concurrency = 1
		}

		for i := 0; i < concurrency; i++ {
			workerID := fmt.Sprintf("%s:worker:%d", processID, i)

			go func(key string, config config.CustomQueue, id string) {
				for {
					worker(key, config, id)
					// Small pause between iterations to prevent tight loop
					time.Sleep(10 * time.Second)
				}
			}(queueKey, queueConfig, workerID)
		}
	}
}
