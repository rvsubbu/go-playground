package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
    "sync/atomic"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
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
        writeJSON(w, http.StatusOK, map[string]string{"version": version})
    })

    r.Get("/config", func(w http.ResponseWriter, _ *http.Request) {
        writeJSON(w, http.StatusOK, config{
            Environment: envOrDefault("APP_ENV", "development"),
            LogLevel:    envOrDefault("LOG_LEVEL", "info"),
        })
    })

    r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
        writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
    })

    r.Get("/metrics", func(w http.ResponseWriter, _ *http.Request) {
        w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
        fmt.Fprintf(w,
            "# HELP http_requests_total Total HTTP requests served.\n"+
                "# TYPE http_requests_total counter\n"+
                "http_requests_total %d\n",
            httpRequestsTotal.Load(),
        )
    })

    server := &http.Server{
        Addr:              ":" + envOrDefault("PORT", "8080"),
        Handler:           r,
        ReadHeaderTimeout: 5 * time.Second,
    }

    log.Printf("API server listening on %s", server.Addr)
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

func writeJSON(w http.ResponseWriter, status int, value any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(value); err != nil {
        log.Printf("write JSON response: %v", err)
    }
}

func envOrDefault(key, fallback string) string {
    if value := strings.TrimSpace(os.Getenv(key)); value != "" {
        return value
    }
    return fallback
}
