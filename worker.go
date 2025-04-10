package main

import (
	"flag"
	"log"
	"os"
	"time"

	cronsched "worker/cronSched"
	db "worker/dbRedis"
	hb "worker/heartbeat"
	"worker/prune"

	"github.com/sevlyar/go-daemon"

	config "worker/config"
)

var (
	cfg *config.Config
	d   = flag.Int("d",
		0,
		"run as daemon, 0 by default , 1 to run as daemon")
	u    = flag.String("u", "redis://localhost:6739", "redis url")
	name = flag.String("name", "worker1", "name of worker")
)

func main() {
	flag.Parse()
	loc, _ := time.LoadLocation("UTC")
	time.Local = loc

	db.RedisUrl = u
	cfg = config.Get()

	if len(os.Args) > 1 {
		cfg.PidFileName = *name + ".pid"
		cfg.LogFileName = *name + ".log"
		cfg.WorkerProcName = *name
	}
	if *d == 1 {
		cntxt := &daemon.Context{
			PidFileName: cfg.PidFileName,
			PidFilePerm: 0o644,
			LogFileName: cfg.LogFileName,
			LogFilePerm: 0o640,
			WorkDir:     "./",
			Umask:       0o27,
			Args:        os.Args,
		}

		d, err := cntxt.Reborn()
		if err != nil {
			log.Fatal("Unable to run: ", err)
		}
		if d != nil {
			return
		}
		defer cntxt.Release()
	}

	log.Println("name is ", cfg.WorkerProcName)

	var forever chan string
	hb.HearBeatStart()
	cronsched.AddPendingCronsStart()
	prune.PruneStart()
	<-forever
}
