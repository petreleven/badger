package main

import (
	"context"
	"flag"
	"log"
	"time"

	listing "worker/cronlisting"

	"github.com/redis/go-redis/v9"
)

var cmd *int = flag.Int("t", 0, "get 0, set 1")
var RedisUrl *string = flag.String("u", "redis://localhost:6739", "redis url")

func main2() {
	flag.Parse()
	op, err := redis.ParseURL(*RedisUrl)
	if err != nil {
		log.Panicln("ERROR parsing redis url")
	}
	options := &redis.Options{
		Addr:             op.Addr,
		Password:         op.Password,
		DB:               1,
		DisableIndentity: true,
		MaxRetries:       10,
		PoolSize:         2,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      10 * time.Second,
		WriteTimeout:     5 * time.Second,
		PoolTimeout:      1 * time.Second,
	}

	if err != nil {
		log.Println("Cant parse url")
		return
	}
	redisClient := redis.NewClient(options)
	/*
	 */
	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalln("Failed to connect to redis: ", err)
	}
	log.Println("Connected to Redis")

	ctx := context.Background()
	c1 := listing.Cron{
		Name:    "start",
		Minute:  "*",
		Hour:    "*/20",
		Day:     "*",
		Month:   "*",
		DayWeek: "*",
		Job:     "run bash",
		Queue:   "userqueue",
	}
	c2 := listing.Cron{
		Name:    "stop",
		Minute:  "20",
		Hour:    "10",
		Day:     "09",
		Month:   "04",
		DayWeek: "2",
		Job:     "run bash",
		Queue:   "userqueue",
	}
	if *cmd == 1 {
		status, err := redisClient.HSet(ctx, "userqueue",
			"badgerWorker:"+c1.Name+":100", c1.Encode()).Result()
		if err != nil {
			log.Println("Unable to save to UserQueue: ", err)
		}
		log.Println(status)
		status, err = redisClient.HSet(ctx, "userqueue",
			"badgerWorker:"+c2.Name+":100", c2.Encode()).Result()
		if err != nil {
			log.Println("Unable to save to UserQueue: ", err)
		}
		log.Println(status)
		log.Println("sent")
	}

	r, err := redisClient.HGetAll(ctx, "userqueue").Result()
	log.Println(r)
}

/*func AddToUserQueue(c *listing.Cron) {
	ctx := context.Background()
	_, err := redisClient.HSet(ctx, cfg2.UserQueue,
		c.Name+":string", c.Encode()).Result()
	if err != nil {
		log.Println("Unable to save to UserQueue: ", err)
	}
}

func tester() {
	log.Println("")
	d := time.Now()
	d.Weekday()
	d.Day()
	data, err := json.Marshal(d)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(data))
}
*/
