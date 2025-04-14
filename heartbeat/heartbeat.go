package heartbeat

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"time"

	"worker/config"
	"worker/dbRedis"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	cfg         *config.Config
)

func HearBeatStart() {
	redisClient = dbRedis.Get()
	cfg = config.Get()
	hearBeat()
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			hearBeat()
		}
	}()
}

func generatePayload() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Unable to determine hostname")
	}
	wmd := &WorkerMetaData{
		externalIp(),
		hostname,
		os.Getpid(),
		time.Now().Unix(),
	}
	data, err := json.Marshal(wmd)
	if err != nil {
		log.Fatalln("Unable to set worker ", err)
	}
	return string(data)
}

func hearBeat() {
	// TODO
	// SAVE WORKER METADATA IN A SERVICE
	payload := generatePayload()
	_, err := redisClient.HSet(context.Background(), cfg.AllWorkers, cfg.WorkerProcName, payload).Result()
	if err != nil {
		log.Println("ERROR REGISTERING WORKER:", err)
	}

	/*store := redisClient.HGetAll(context.Background(), "allworkers").Val()
	log.Println("store has")
	log.Println(store)*/
}

func externalIp() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		// interface is down
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		// skip loopback interfaces
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addresses, err := iface.Addrs()
		if err != nil {
			return ""
		}

		for _, addr := range addresses {
			var ip net.IP
			switch t := addr.(type) {
			case *net.IPNet:
				ip = t.IP
			case *net.IPAddr:
				ip = t.IP
			}

			if ip == nil {
				continue
			}
			ip = ip.To4()
			// ip is not ipv4

			if ip == nil || ip.IsLoopback() {
				continue
			}
			return ip.String()
		}
	}
	return ""
}
