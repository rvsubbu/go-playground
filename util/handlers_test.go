package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestVersionHandler(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "1.2.3",
		GitSHA:    "abc123",
		BuildTime: "2026-11-23T12:00:00Z",
	}
	response := httptest.NewRecorder()

	VersionHandler(&buildInfo).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/version", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var got map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := map[string]string{
		"version":    buildInfo.Version,
		"git_sha":    buildInfo.GitSHA,
		"build_time": buildInfo.BuildTime,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("response = %#v, want %#v", got, want)
	}
}

func TestConfigHandler(t *testing.T) {
	previousConfig := BasicConf
	t.Cleanup(func() {
		BasicConf = previousConfig
	})
	BasicConf = &BasicConfig{
		App:    BasicAppConfig{Name: "test-app", Env: "staging"},
		Log:    BasicLogConfig{Level: "info", Requests: true},
		Server: BasicServerConfig{Addr: ":8081", ReadTimeout: 2 * time.Second, WebServer: "chi"},
	}

	response := httptest.NewRecorder()

	ConfigHandler(&BuildInfo{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/config", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got BasicConfig
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := *BasicConf
	if !reflect.DeepEqual(got, want) {
		t.Errorf("response = %#v, want %#v", got, want)
	}
}

func TestHealthzHandler(t *testing.T) {
	response := httptest.NewRecorder()

	HealthzHandler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["status"] != "ok" {
		t.Errorf("status field = %q, want ok", got["status"])
	}
}

func TestMetricsHandler(t *testing.T) {
	response := httptest.NewRecorder()
	wantRequestsTotal := fmt.Sprintf(
		"# HELP http_requests_total Total HTTP requests served.\n"+
			"# TYPE http_requests_total counter\n"+
			"http_requests_total %d\n",
		HttpRequestsTotal.Load(),
	)

	MetricsHandler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "text/plain; version=0.0.4; charset=utf-8" {
		t.Errorf("Content-Type = %q, want Prometheus text format", got)
	}
	body := response.Body.String()
	for _, want := range []string{
		wantRequestsTotal,
		"# TYPE http_request_duration_seconds histogram\n",
		"http_request_duration_seconds_bucket{le=\"+Inf\"}",
		"http_request_duration_seconds_sum ",
		"http_request_duration_seconds_count ",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics body %q does not contain %q", body, want)
		}
	}
}

func TestEnvOrDefault(t *testing.T) {
	const key = "GO_PLAYGROUND_TEST_SETTING"
	t.Setenv(key, "  configured  ")
	if got := EnvOrDefault(key, "fallback"); got != "configured" {
		t.Errorf("EnvOrDefault() = %q, want configured", got)
	}

	t.Setenv(key, " ")
	if got := EnvOrDefault(key, "fallback"); got != "fallback" {
		t.Errorf("EnvOrDefault() with blank value = %q, want fallback", got)
	}
}

func TestIncrRequests(t *testing.T) {
	before := HttpRequestsTotal.Load()
	beforeDurationCount := requestDurationCount.Load()
	beforeDurationSum := requestDurationSumNanos.Load()
	called := false
	handler := IncrRequests(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		time.Sleep(30 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if !called {
		t.Fatal("wrapped handler was not called")
	}
	if response.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := HttpRequestsTotal.Load(); got != before+1 {
		t.Errorf("request count = %d, want %d", got, before+1)
	}
	if got := requestDurationCount.Load(); got != beforeDurationCount+1 {
		t.Errorf("duration observation count = %d, want %d", got, beforeDurationCount+1)
	}
	if got := requestDurationSumNanos.Load() - beforeDurationSum; got < uint64(20*time.Millisecond) {
		t.Errorf("recorded duration = %s, want at least 20ms", time.Duration(got))
	}
}

func TestRecordRequestDurationUpdatesCumulativeBuckets(t *testing.T) {
	beforeCount := requestDurationCount.Load()
	beforeBuckets := make([]uint64, len(requestDurationBuckets))
	for i := range requestDurationBuckets {
		beforeBuckets[i] = requestDurationBucketCounts[i].Load()
	}

	recordRequestDuration(25 * time.Millisecond)

	if got := requestDurationCount.Load(); got != beforeCount+1 {
		t.Fatalf("duration count = %d, want %d", got, beforeCount+1)
	}
	for i, bucket := range requestDurationBuckets {
		want := beforeBuckets[i]
		if bucket.max >= 25*time.Millisecond {
			want++
		}
		if got := requestDurationBucketCounts[i].Load(); got != want {
			t.Errorf("bucket le=%s count = %d, want %d", bucket.label, got, want)
		}
	}
}
