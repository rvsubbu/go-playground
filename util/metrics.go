package util

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	requestDurationBuckets = [...]struct {
		label string
		max   time.Duration
	}{
		{label: "0.005", max: 5 * time.Millisecond},
		{label: "0.01", max: 10 * time.Millisecond},
		{label: "0.025", max: 25 * time.Millisecond},
		{label: "0.05", max: 50 * time.Millisecond},
		{label: "0.1", max: 100 * time.Millisecond},
		{label: "0.25", max: 250 * time.Millisecond},
		{label: "0.5", max: 500 * time.Millisecond},
		{label: "1", max: time.Second},
		{label: "2.5", max: 2500 * time.Millisecond},
		{label: "5", max: 5 * time.Second},
		{label: "10", max: 10 * time.Second},
	}

	requestDurationBucketCounts [len(requestDurationBuckets)]atomic.Uint64
	requestDurationCount        atomic.Uint64
	requestDurationSumNanos     atomic.Uint64

	requestCounters = struct {
		sync.RWMutex
		counts map[requestCounterLabels]uint64
	}{
		counts: make(map[requestCounterLabels]uint64),
	}
)

type requestCounterLabels struct {
	appName     string
	method      string
	status      int
	apiNickname string
}

type requestCounterSample struct {
	labels requestCounterLabels
	count  uint64
}

type responseStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *responseStatusRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseStatusRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (w *responseStatusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func InstrumentRequests(appName string) func(http.Handler) http.Handler {
	if appName == "" {
		appName = "unknown"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &responseStatusRecorder{ResponseWriter: w}
			slog.Debug("Request starts", "method", r.Method, "path", r.URL.Path)
			defer func() {
				duration := time.Since(start)
				recovered := recover()
				status := recorder.status
				if status == 0 {
					status = http.StatusOK
				}
				if recovered != nil && recorder.status == 0 {
					status = http.StatusInternalServerError
				}
				slog.Debug("Request ends",
					"method", r.Method,
					"path", r.URL.Path,
					"status", status,
					"time_taken", duration,
				)
				recordRequestCounter(appName, r.Method, status, apiNickname(r.URL.Path))
				recordRequestDuration(duration)
				if recovered != nil {
					panic(recovered)
				}
			}()
			next.ServeHTTP(recorder, r)
		})
	}
}

func recordRequestDuration(duration time.Duration) {
	if duration < 0 {
		duration = 0
	}

	requestDurationCount.Add(1)
	requestDurationSumNanos.Add(uint64(duration))
	for i, bucket := range requestDurationBuckets {
		if duration <= bucket.max {
			requestDurationBucketCounts[i].Add(1)
		}
	}
}

func recordRequestCounter(appName, method string, status int, apiNickname string) {
	labels := requestCounterLabels{
		appName:     appName,
		method:      method,
		status:      status,
		apiNickname: apiNickname,
	}
	requestCounters.Lock()
	requestCounters.counts[labels]++
	requestCounters.Unlock()
}

func requestCounterSnapshot() []requestCounterSample {
	requestCounters.RLock()
	defer requestCounters.RUnlock()

	samples := make([]requestCounterSample, 0, len(requestCounters.counts))
	for labels, count := range requestCounters.counts {
		samples = append(samples, requestCounterSample{labels: labels, count: count})
	}
	return samples
}

func apiNickname(path string) string {
	switch strings.Trim(path, "/") {
	case "version", "config", "healthz", "metrics":
		return strings.Trim(path, "/")
	default:
		return "unknown"
	}
}

func escapePrometheusLabel(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\"", "\\\"")
	return `"` + replacer.Replace(value) + `"`
}

func formatStatus(status int) string {
	return strconv.Itoa(status)
}
