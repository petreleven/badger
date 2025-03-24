package cronlisting

import (
	"context"
	"log"
	"strings"

	"worker/config"
	"worker/dbRedis"
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
		return nil, err
	}
	//[startserver]10,11,12,13SystemspecsUseroptions
	userCrons := []Cron{}
	for cronName, cronData := range results {
		var newCron Cron
		// expectation min hour day/date month
		cronDetails := strings.Fields(cronData)
		// TODO DIFFERENT QUEUES FOR EACH CRON
		err = newCron.DecodeFromSlice(cronName, cronDetails)
		if err != nil {
			log.Println("Error decoding cron :", err)
			continue
		}
		userCrons = append(userCrons, newCron)
	}
	return &userCrons, nil
}
