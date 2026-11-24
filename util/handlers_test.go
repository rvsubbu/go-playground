package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestVersionHandler(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "1.2.3",
		GitSHA:    "abc123",
		BuildTime: "2026-11-23T12:00:00Z",
		Env:       "test",
		LogLevel:  "debug",
	}
	response := httptest.NewRecorder()

	VersionHandler(buildInfo).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/version", nil))

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
		"env":        buildInfo.Env,
		"log_level":  buildInfo.LogLevel,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("response = %#v, want %#v", got, want)
	}
}

func TestConfigHandler(t *testing.T) {
	response := httptest.NewRecorder()
	buildInfo := BuildInfo{Env: "staging", LogLevel: "info"}

	ConfigHandler(buildInfo).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/config", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	var got map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := map[string]string{"env": "staging", "log_level": "info"}
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
	want := fmt.Sprintf(
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
	if got := response.Body.String(); got != want {
		t.Errorf("body = %q, want %q", got, want)
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
	called := false
	handler := IncrRequests(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
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
}
