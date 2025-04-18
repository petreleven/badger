package web

import (
	"context"
	"encoding/json"
	"html/template"
	"maps"
	"net/http"
	"path/filepath"
	"strings"

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

func showQueuePreview(w http.ResponseWriter, req *http.Request) {
	var (
		cfg         = config.Get()
		redisClient = db.Get()
		ctx         = context.Background()
	)
	queueNames := maps.Keys(cfg.CustomQueues.Queues)

	type singleQueueStruct struct {
		Name        string
		Concurrency int
		PendingLen  int64
		RunningLen  int64
		DelayedLen  int64
		FailedLen   int64
		DoneLen     int64
	}
	data := struct {
		AllQueues []singleQueueStruct
	}{
		[]singleQueueStruct{},
	}

	for queueKey := range queueNames {
		singleQueue := singleQueueStruct{
			Name:        queueKey,
			Concurrency: cfg.CustomQueues.Queues[queueKey].Concurrency,
		}
		pendingQueue := "badger:pending:" + queueKey
		runningHash := "badger:running:" + queueKey
		delayedQueue := "badger:delayed:" + queueKey
		failedQueue := "badger:failed:" + queueKey
		doneQueue := "badger:done:" + queueKey

		pendingLen, _ := redisClient.LLen(ctx, pendingQueue).Result()
		runningLen, _ := redisClient.HLen(ctx, runningHash).Result()
		delayedLen, _ := redisClient.LLen(ctx, delayedQueue).Result()
		failedLen, _ := redisClient.LLen(ctx, failedQueue).Result()
		doneLen, _ := redisClient.LLen(ctx, doneQueue).Result()

		singleQueue.PendingLen = pendingLen
		singleQueue.RunningLen = runningLen
		singleQueue.DelayedLen = delayedLen
		singleQueue.FailedLen = failedLen
		singleQueue.DoneLen = doneLen
		data.AllQueues = append(data.AllQueues, singleQueue)
	}
	path := filepath.Join(templateAbs, "jobs.html")
	tmpl, _ := template.ParseFiles(path)
	tmpl = template.Must(tmpl, nil)
	tmpl.Execute(w, data)
}

func inspectQueue(w http.ResponseWriter, req *http.Request) {
	var (
		redisClient = db.Get()
		ctx         = context.Background()
	)
	htmxHeader := req.Header.Get("Hx-Request")
	templateName := ""
	if htmxHeader == "" {
		templateName = "inspectQueueFull.html"
	} else {
		templateName = "inspectQueue.html"
	}

	queueName := req.URL.Query().Get("queuename")
	data := struct {
		Jobs []string
	}{Jobs: []string{}}
	if strings.HasPrefix(queueName, "badger:running") {
		res, _ := redisClient.HGetAll(ctx, queueName).Result()
		ks := maps.Keys(res)
		for k := range ks {
			data.Jobs = append(data.Jobs, k)
		}
	} else {
		res, _ := redisClient.LRange(ctx, queueName, 0, 50).Result()
		data.Jobs = res
	}
	path := filepath.Join(templateAbs, templateName)
	tmpl, _ := template.ParseFiles(path)
	tmpl = template.Must(tmpl, nil)
	tmpl.Execute(w, data)
}
