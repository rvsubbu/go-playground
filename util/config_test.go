package util

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestConfigInitLoadsYAML(t *testing.T) {
	restoreBasicConfig(t)
	configFile := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`
app:
  name: test-api
  env: staging
logging:
  level: warn
  requests: false
server:
  addr: ":9090"
  read_timeout: 3s
  web_server: chi
`)
	if err := os.WriteFile(configFile, data, 0600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if err := ConfigInit(configFile); err != nil {
		t.Fatalf("ConfigInit() error = %v", err)
	}

	want := &BasicConfig{
		App:    BasicAppConfig{Name: "test-api", Env: "staging"},
		Log:    BasicLogConfig{Level: "warn", Requests: false},
		Server: BasicServerConfig{Addr: ":9090", ReadTimeout: 3 * time.Second, WebServer: "chi"},
	}
	if !reflect.DeepEqual(BasicConf, want) {
		t.Errorf("BasicConf = %#v, want %#v", BasicConf, want)
	}
}

func TestConfigInitPreservesDefaults(t *testing.T) {
	restoreBasicConfig(t)
	configFile := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configFile, []byte("server:\n  web_server: chi\n"), 0600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	if err := ConfigInit(configFile); err != nil {
		t.Fatalf("ConfigInit() error = %v", err)
	}

	want := &BasicConfig{
		App:    BasicAppConfig{Name: "unknown", Env: "dev"},
		Log:    BasicLogConfig{Level: "debug", Requests: true},
		Server: BasicServerConfig{Addr: ":8080", ReadTimeout: time.Second, WebServer: "chi"},
	}
	if !reflect.DeepEqual(BasicConf, want) {
		t.Errorf("BasicConf = %#v, want %#v", BasicConf, want)
	}
}

func TestConfigInitReturnsFileError(t *testing.T) {
	restoreBasicConfig(t)
	configFile := filepath.Join(t.TempDir(), "missing.yaml")

	if err := ConfigInit(configFile); err == nil {
		t.Fatal("ConfigInit() error = nil, want file error")
	}
}

func restoreBasicConfig(t *testing.T) {
	t.Helper()
	previousConfig := BasicConf
	t.Cleanup(func() {
		BasicConf = previousConfig
	})
}
