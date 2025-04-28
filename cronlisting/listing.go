package cronlisting

import (
	"context"
	"log"
	"strings"

	"worker/dbRedis"
)

func GetQueuedTasks(queueName string) (*[]Cron, error) {
	redisClient := dbRedis.Get()
	ctx := context.Background()
	results, err := redisClient.HGetAll(ctx, queueName).Result()
	if err != nil {
		log.Println("Error getting ", queueName)
		return nil, err
	}
	//[startserver]10,11,12,13SystemspecsUseroptions
	s := []Cron{}
	var  userCrons *[]Cron = &s
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
		*userCrons = append(*userCrons, newCron)
	}
	return userCrons, nil
}
