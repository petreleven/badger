package web

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"

	"worker/config"
	db "worker/dbRedis"
	hb "worker/heartbeat"
)

type allworkersStruct struct {
	listofworkers []hb.WorkerMetaData
}

func homepage(w http.ResponseWriter, req *http.Request) {
	path := filepath.Join(templateAbs, "home.html")
	t, err := template.ParseFiles(path)
	t = template.Must(t, nil)
	if err != nil {
		errlogger(err)
		return
	}
	data := struct {
		CustomQueues []string
	}{
		CustomQueues: []string{},
	}
	cfg := config.Get()
	for key := range cfg.CustomQueues.Queues {
		data.CustomQueues = append(data.CustomQueues, key)
	}

	t.Execute(w, data)
}

func getWorkers(w http.ResponseWriter, req *http.Request) {
	var (
		cfg         = config.Get()
		redisClient = db.Get()
	)
	type singleWorkerData struct {
		WorkerName string
		HbMetaData hb.WorkerMetaData
	}

	ctx := context.Background()
	allworkers, _ := redisClient.HGetAll(ctx, cfg.ClusterName).Result()
	renderData := []singleWorkerData{}
	for key, value := range allworkers {
		data := singleWorkerData{WorkerName: key}
		json.Unmarshal([]byte(value), &data.HbMetaData)
		renderData = append(renderData, data)
	}

	path := filepath.Join(templateAbs, "allworkers.html")
	t, err := template.ParseFiles(path)
	if err != nil {
		errlogger(err)
	}
	renderDataStruct := struct {
		Name    string
		Workers []singleWorkerData
	}{
		Name:    "Workers",
		Workers: renderData,
	}
	t = template.Must(t, nil)
	t.Execute(w, renderDataStruct)
}
