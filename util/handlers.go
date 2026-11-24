package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type BuildInfo struct {
	Version   string `json: version`
	GitSHA    string `json: git_sha`
	BuildTime string `json: build_time`
	Env       string `json: env`
	LogLevel  string `json: log_level`
}

func VersionHandler(buildInfo BuildInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSONResponse(w, http.StatusOK, buildInfo)
	}
}

func ConfigHandler(buildInfo BuildInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config := map[string]string{"Env": buildInfo.Env, "LogLevel": buildInfo.LogLevel,}
		WriteJSONResponse(w, http.StatusOK, config)
	}
}

func HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	    w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	    fmt.Fprintf(w,
	        "# HELP http_requests_total Total HTTP requests served.\n"+
	            "# TYPE http_requests_total counter\n"+
	            "http_requests_total %d\n",
	        HttpRequestsTotal.Load(),
		)
	}
}

func WriteJSONResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}

func EnvOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

/*
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
)


type config struct {
	Environment string `json:"environment"`
	LogLevel    string `json:"logLevel"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /version" func(w http.ResponseWriter, _*http.Request) {
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/version", getOnly(func(w http.ResponseWriter, _ *http.Request) {
	    writeJSON(w, http.StatusOK, map[string]string{"version": version})
	}))

	mux.HandleFunc("/config", getOnly(func(w http.ResponseWriter, _ *http.Request) {
	    writeJSON(w, http.StatusOK, config{
	        Environment: envOrDefault("APP_ENV", "development"),
	        LogLevel:    envOrDefault("LOG_LEVEL", "info"),
	    })
	}))

	mux.HandleFunc("/healthz", getOnly(func(w http.ResponseWriter, _ *http.Request) {
	    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}))

	mux.HandleFunc("/metrics", getOnly(func(w http.ResponseWriter, _ *http.Request) {
	    w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	    fmt.Fprintf(w,
	        "# HELP http_requests_total Total HTTP requests served.\n"+
	            "# TYPE http_requests_total counter\n"+
	            "http_requests_total %d\n",
	        httpRequestsTotal.Load(),
	    )
	}))

	server := &http.Server{
	    Addr:              ":" + envOrDefault("PORT", "8080"),
	    Handler:           countRequests(mux),
	    ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	    log.Fatal(err)
	}
}
*/
