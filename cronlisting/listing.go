package cronlisting

import (
	"context"
	"log"
	"strings"

	"worker/config"
	"worker/dbRedis"

	"github.com/google/uuid"
)

var (
	redisClient               = dbRedis.Get()
	cfg         config.Config = *config.Get()
)

func GetUserQueuedTasks() (*[]Cron, error) {
	ctx := context.Background()
	results, err := redisClient.HGetAll(ctx, cfg.UserQueue).Result()
	if err != nil {
		log.Println("Error getting UserQueue")
		return nil, nil
	}
	//[startserver]10,11,12,13SystemspecsUseroptions
	var userCrons []Cron
	for cronName, cronData := range results {
		var newCron Cron
		newCron.Name = cronName
		// expectation min hour day/date month
		cronDetails := strings.Fields(cronData)
		// TODO DIFFERENT QUEUES FOR EACH CRON
		newCron = Cron{
			Name:   cronName,
			Minute: cronDetails[0],
			Hour:   cronDetails[1],
			Day:    cronDetails[2],
			Month:  cronDetails[3],
			Job:    "do this",
			Queue:  cfg.PendingQueue,
		}
		userCrons = append(userCrons, newCron)
	}
	return &userCrons, nil
}

func AddToUserQueue() {
	ctx := context.Background()
	c := Cron{
		Name:   "test",
		Minute: "00",
		Hour:   "01",
		Day:    "29",
		Month:  "12",
	}
	_, err := redisClient.HSet(ctx, cfg.UserQueue, c.Name+":"+uuid.NewString(), c.Encode()).Result()
	if err != nil {
		log.Println("Unable to save to UserQueue: ", err)
	}
}
