package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

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

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(util.IncrRequests)

	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		util.VersionHandler(buildInfo)(w, r)
	})
	r.Get("/config", func(w http.ResponseWriter, r *http.Request) {
		util.ConfigHandler(buildInfo)(w, r)
	})
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		util.HealthzHandler()(w, r)
	})
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		util.MetricsHandler()(w, r)
	})

	server := &http.Server{
		Addr: ":" + util.EnvOrDefault("PORT", "8080"),
		Handler: r,
		ReadHeaderTimeout: 1 * time.Second,
	}
	log.Printf("Basic API Server listening on port %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
