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

var (
	cfg              *config.Config = config.Get()
	procName         *string        = flag.String("n", "", "The worker name")
	g_workerProcName string
)

func main() {
	flag.Parse()

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
