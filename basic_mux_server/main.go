package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/rvsubbu/go-playground/util"
)

var (
	version   = "0.1.0"
	gitSHA    = "unknown"
	buildTime = "unknown"
	env       = "dev"
)

func main() {
	buildInfo := util.BuildInfo{
    	Version:   version,
    	GitSHA:    gitSHA,
    	BuildTime: buildTime,
    	Env:       env,
		LogLevel:  "info",
	}
	if buildInfo.Env == "dev" {
		buildInfo.LogLevel = "debug"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		util.VersionHandler(buildInfo)(w, r)
	})

	mux.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		util.ConfigHandler(buildInfo)(w, r)
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		util.HealthzHandler()(w, r)
	})

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		util.MetricsHandler()(w, r)
	})

	server := &http.Server{
		Addr:              ":" + util.EnvOrDefault("PORT", "8080"),
		Handler:           util.IncrRequests(mux),
		ReadHeaderTimeout: 1 * time.Second,
	}
	log.Printf("API server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	    log.Fatal(err)
	}
}
