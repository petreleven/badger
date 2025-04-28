package web

import (
	"context"
	"encoding/json"
	"html/template"
	"maps"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"worker/config"
	db "worker/dbRedis"
	hb "worker/heartbeat"
)

type allworkersStruct struct {
	listofworkers []hb.WorkerMetaData
}

func homepage(w http.ResponseWriter, req *http.Request) {
	var (
		cfg         = config.Get()
		redisClient = db.Get()
		ctx         = context.Background()
	)
	path := filepath.Join(templateAbs, "home.html")
	t, err := template.ParseFiles(path)
	t = template.Must(t, nil)

	if err != nil {
		errlogger(err)
		return
	}
	data := struct {
		CustomQueues []string
		WorkersCount int64
		RunningCount int64
		DoneCount    int64
		FailedCount  int64
	}{
		CustomQueues: []string{},
	}
	workers, _ := redisClient.HLen(ctx, cfg.ClusterName).Result()
	data.WorkersCount = workers
	for key := range cfg.CustomQueues.Queues {
		data.CustomQueues = append(data.CustomQueues, key)

		runningHash := "badger:running:" + key
		failedQueue := "badger:failed:" + key
		doneQueue := "badger:done:" + key
		runningLen, _ := redisClient.HLen(ctx, runningHash).Result()
		failedLen, _ := redisClient.LLen(ctx, failedQueue).Result()
		doneLen, _ := redisClient.LLen(ctx, doneQueue).Result()
		data.RunningCount += runningLen
		data.DoneCount += doneLen
		data.FailedCount += failedLen
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

	var tmpl *template.Template
	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"div": func(i int64, j int64) int {
			res := int(i / j)
			if res <= 0 {
				res = 1
			}
			return res
		},
	}

	// check htmx headers
	if req.Header.Get("Hx-Request") == "" {
		path1 := filepath.Join(templateAbs, "inspectQueueFull.html")
		path2 := filepath.Join(templateAbs, "inspectQueue.html")
		tmpl = template.Must(
			template.New("inspectQueueFull.html").
				Funcs(funcMap).
				ParseFiles(path1, path2))

	} else {
		path1 := filepath.Join(templateAbs, "inspectQueue.html")
		tmpl = template.Must(
			template.New("inspectQueue.html").
				Funcs(funcMap).
				ParseFiles(path1))
	}

	// get queries
	queueName := req.URL.Query().Get("queuename")
	startStr := req.URL.Query().Get("start")
	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		start = 0
	}
	var jobsRange int64 = 10
	stop := start + jobsRange

	data := struct {
		Jobs      []string
		JobsID    []string
		Total     int64
		Name      string
		PrevStart int64
		NextStart int64
	}{Jobs: []string{}}
	if strings.HasPrefix(queueName, "badger:running") {
		res, _ := redisClient.HGetAll(ctx, queueName).Result()
		ks := maps.Keys(res)
		for k := range ks {
			data.Jobs = append(data.Jobs, k)
		}
		data.Total, _ = redisClient.HLen(ctx, queueName).Result()
	} else {
		res, _ := redisClient.LRange(ctx, queueName, start, stop).Result()

		for _, v := range res {

			s := strings.Split(v, ":")
			if len(s) > 1 {
				id := s[0]
				job := s[1]
				data.Jobs = append(data.Jobs, job)
				data.JobsID = append(data.JobsID, id)
			} else {
				data.Jobs = append(data.Jobs, v)
				data.JobsID = append(data.JobsID, "")
			}

		}
		data.Total, _ = redisClient.LLen(ctx, queueName).Result()
	}
	data.Name = queueName
	data.NextStart = stop
	data.PrevStart = start - jobsRange
	if data.PrevStart < 0 {
		data.PrevStart = 0
	}
	tmpl.Execute(w, data)
}

func inspectJob(w http.ResponseWriter, req *http.Request) {
	var (
		redisClient = db.Get()
		ctx         = context.Background()
	)
	htmxHeader := req.Header.Get("Hx-Request")
	logid := req.URL.Query().Get("logid")
	var tmpl *template.Template
	if htmxHeader == "" {
		path1 := filepath.Join(templateAbs, "joblogsfull.html")
		path2 := filepath.Join(templateAbs, "joblogs.html")
		tmpl, _ = template.ParseFiles(path1, path2)
		tmpl = template.Must(tmpl, nil)
	} else {
		path := filepath.Join(templateAbs, "joblogs.html")
		tmpl, _ = template.ParseFiles(path)
	}

	res, _ := redisClient.HGet(ctx, "badger:joblog", logid).Result()
	data := struct {
		Logs string
	}{}
	data.Logs = res
	if data.Logs == "" {
		data.Logs = "Logging is set to false in config file"
	}
	tmpl.Execute(w, data)
}

func requeueOrDelete(w http.ResponseWriter, req *http.Request) {
	var (
		redisClient = db.Get()
		ctx         = context.Background()
	)
	jobId := req.URL.Query().Get("jobid")
	job := req.URL.Query().Get("job")
	queueName := req.URL.Query().Get("queuename")
	operation := req.URL.Query().Get("operation")
	if operation == "delete" {
		redisClient.LRem(ctx, queueName, 1, jobId+":"+job)
	} else {
		s := strings.Split(queueName, ":")
		if len(s) == 3 {
			pendingQueue := s[0] + ":pending:" + s[2]
			redisClient.LPush(ctx, pendingQueue, job)
		}
	}

	redirectUrl := "/inspectQueue?start=0" + "&queuename=" + queueName
	http.Redirect(w, req, redirectUrl, http.StatusMovedPermanently)
}
