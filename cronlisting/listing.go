package cronlisting

import (
	"context"
	"encoding/json"
	"log"

	"worker/dbRedis"
)

func GetQueuedTasks(queueName string) (*[]Cron, error) {
	redisClient := dbRedis.Get()
	ctx := context.Background()
	results, err := redisClient.HGetAll(ctx, queueName).Result()
	log.Printf("HGETALL queue:%s yielded %v\n", queueName, len(results))
	if err != nil {
		log.Println("Error getting ", queueName)
		return nil, err
	}
	//[startserver]10,11,12,13SystemspecsUseroptions
	s := []Cron{}
	var userCrons *[]Cron = &s
	for _, cronData := range results {
		var newCron Cron
		err := json.Unmarshal([]byte(cronData), &newCron)
		if err != nil {
			log.Println("Error decoding cron :", err)
			continue
		}
		*userCrons = append(*userCrons, newCron)
	}
	return userCrons, nil
}
