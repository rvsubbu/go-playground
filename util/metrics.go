package util

import (
	"net/http"
	"sync/atomic"
	"time"
)

var (
	HttpRequestsTotal atomic.Uint64

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
)

func IncrRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		HttpRequestsTotal.Add(1)
		defer func() {
			recordRequestDuration(time.Since(start))
		}()
		next.ServeHTTP(w, r)
	})
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
