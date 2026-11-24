package util

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Priority: cmdline flags>env vars>config>defaults
// cmdline flag and env vars not needed for now

type BasicAppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type BasicLogConfig struct {
	Level    string `yaml:"level"`
	Requests bool   `yaml:"requests"`
	// Future
	//Db       bool `yaml: "db"`
	//Tracing  bool `yaml: "tracing"`
	//Kafka    bool `yaml: "kafka"`
}

type BasicServerConfig struct {
	Addr        string        `yaml:"addr"`
	ReadTimeout time.Duration `yaml:"read_timeout"`
	WebServer   string        `yaml:"web_server"`
}

type BasicConfig struct {
	App    BasicAppConfig    `yaml:"app"`
	Log    BasicLogConfig    `yaml:"logging"`
	Server BasicServerConfig `yaml:"server"`
}

var BasicConf *BasicConfig

func defaultBasicConfig() {
	BasicConf = &BasicConfig{
		App:    BasicAppConfig{Name: "unknown", Env: "dev"},
		Log:    BasicLogConfig{Level: "debug", Requests: true},
		Server: BasicServerConfig{Addr: ":8080", ReadTimeout: 1 * time.Second},
	}
}

func ConfigInit(configFile string) error {
	defaultBasicConfig()

	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, BasicConf); err != nil {
		return err
	}
	return nil
}
