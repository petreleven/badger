package main

import (
	"flag"
	"log"
	"os"

	cronsched "worker/cronSched"
	hb "worker/heartbeat"

	"github.com/sevlyar/go-daemon"

	config "worker/config"
)

var cfg *config.Config = config.Get()

func main2() {
	flag.Parse()

	if len(os.Args) > 1 {
		cfg.PidFileName = os.Args[1] + ".pid"
		cfg.LogFileName = os.Args[1] + ".log"
		cfg.WorkerProcName = os.Args[1]
	}
	if true {
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
	<-forever
}
