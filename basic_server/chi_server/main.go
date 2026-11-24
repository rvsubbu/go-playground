package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/rvsubbu/go-playground/util"
)

var (
	version           = "0.1.0"
	httpRequestsTotal atomic.Uint64
)

type config struct {
	Environment string `json:"environment"`
	LogLevel    string `json:"logLevel"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(countRequests)

	r.Get("/version", func(w http.ResponseWriter, _ *http.Request) {
		util.WriteJSONResponse(w, http.StatusOK, map[string]string{"version": version})
	})
	r.Get("/config", func(w http.ResponseWriter, _ *http.Request) {
		util.WriteJSONResponse(w, http.StatusOK, config{
			Environment: util.EnvOrDefault("APP_ENV", "development"),
			LogLevel:    util.EnvOrDefault("LOG_LEVEL", "debug"),
		})
	})
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		util.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintf(w,
			"# HELP http_requests_total Total HTTP Requests served.\n"+
				"# TYPE http_requests_total counter\n"+
				"http_requests_total %d\n", httpRequestsTotal.Load(),)
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

func countRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpRequestsTotal.Add(1)
		next.ServeHTTP(w, r)
	})
}
