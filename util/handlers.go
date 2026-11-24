package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

type BuildInfo struct {
	Version   string `json:"version"`
	GitSHA    string `json:"git_sha"`
	BuildTime string `json:"build_time"`
}

func VersionHandler(buildInfo *BuildInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSONResponse(w, http.StatusOK, buildInfo)
	}
}

func ConfigHandler(buildInfo *BuildInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteJSONResponse(w, http.StatusOK, BasicConf)
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
		fmt.Fprintln(w,
			"# HELP http_requests_total Total HTTP requests served.\n"+
				"# TYPE http_requests_total counter",
		)
		samples := requestCounterSnapshot()
		sort.Slice(samples, func(i, j int) bool {
			a, b := samples[i].labels, samples[j].labels
			if a.appName != b.appName {
				return a.appName < b.appName
			}
			if a.method != b.method {
				return a.method < b.method
			}
			if a.status != b.status {
				return a.status < b.status
			}
			return a.apiNickname < b.apiNickname
		})
		for _, sample := range samples {
			labels := sample.labels
			fmt.Fprintf(w,
				"http_requests_total{app_name=%s,method=%s,status=%s,apiNickname=%s} %d\n",
				escapePrometheusLabel(labels.appName),
				escapePrometheusLabel(labels.method),
				escapePrometheusLabel(formatStatus(labels.status)),
				escapePrometheusLabel(labels.apiNickname),
				sample.count,
			)
		}
		fmt.Fprintln(w,
			"# HELP http_request_duration_seconds HTTP request latency in seconds.\n"+
				"# TYPE http_request_duration_seconds histogram",
		)
		for i, bucket := range requestDurationBuckets {
			fmt.Fprintf(w, "http_request_duration_seconds_bucket{le=%q} %d\n",
				bucket.label, requestDurationBucketCounts[i].Load())
		}
		count := requestDurationCount.Load()
		fmt.Fprintf(w, "http_request_duration_seconds_bucket{le=\"+Inf\"} %d\n", count)
		fmt.Fprintf(w, "http_request_duration_seconds_sum %.9f\n",
			float64(requestDurationSumNanos.Load())/float64(1_000_000_000))
		fmt.Fprintf(w, "http_request_duration_seconds_count %d\n", count)
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
