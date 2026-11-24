package util

import (
	"net/http"
	"sync/atomic"
)

var (
	HttpRequestsTotal atomic.Uint64
)

func IncrRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    HttpRequestsTotal.Add(1)
	    next.ServeHTTP(w, r)
	})
}
