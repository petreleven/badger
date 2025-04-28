package web

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

const (
	keyServeraddr = "serverAddr"
	templateDir   = "web/templates"
)

var (
	cwd         string
	templateAbs string
)

func errlogger(err error) {
	log.Println("unexpected Error Occurred:", err)
}

func WebStart() {
	cwd, _ = os.Getwd()
	templateAbs = filepath.Join(cwd, templateDir)

	mux := http.NewServeMux()

	mux.HandleFunc("/", homepage)
	mux.HandleFunc("/allworkers", getWorkers)
	mux.HandleFunc("/showQueuePreview", showQueuePreview)
	mux.HandleFunc("/inspectQueue", inspectQueue)
	mux.HandleFunc("/inspectJob", inspectJob)
	mux.HandleFunc("/requeueOrDelete", requeueOrDelete)

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()
	server := &http.Server{
		Addr:    "127.0.0.1:5000",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx := context.WithValue(ctx, keyServeraddr, l.Addr().String())
			return ctx
		},
	}
	log.Println("Starting Web Server port 5000")
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("Server Closed")
	} else {
		errlogger(err)
	}
}
