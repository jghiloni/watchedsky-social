package config

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	Host                   string `yaml:"host" envconfig:"host"`
	Username               string `yaml:"username" envconfig:"username"`
	Password               string `yaml:"password" envconfig:"password"`
	Name                   string `yaml:"database" envconfig:"database"`
	AuthenticationDatabase string `yaml:"authdb" envconfig:"authdb"`
}

type BlueskyConfig struct {
	PDSURL         string `yaml:"pds_url" envconfig:"pds_url"`
	Username       string `yaml:"username" envconfig:"username"`
	AppPassword    string `yaml:"app_password" envconfig:"app_password"`
	FeedServiceDID string `yaml:"feed_service_did" envconfig:"feed_service_did"`
}

type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled" envconfig:"enabled"`
	Port    uint16 `yaml:"port" envconfig:"port"`
}

type HTTPServerConfig struct {
	Enabled   bool   `yaml:"enabled" envconfig:"enabled"`
	Port      uint16 `yaml:"port" envconfig:"port"`
	StorageDB string `yaml:"storage_db" envconfig:"storage_db"`
}

type AlertPollConfig struct {
	Enabled      bool          `yaml:"enabled" envconfig:"enabled"`
	PollInterval time.Duration `yaml:"poll_interval" envconfig:"poll_interval"`
}

type FirehoseConfig struct {
	Enabled bool `yaml:"enabled" envconfig:"enabled"`
}

type AppConfig struct {
	BaseURL        string           `yaml:"base_url" envconfig:"base_url"`
	LogLevel       LogLevel         `yaml:"log_level" envconfig:"log_level"`
	MongoDB        DatabaseConfig   `yaml:"database" envconfig:"database"`
	Bluesky        BlueskyConfig    `yaml:"bluesky" envconfig:"bluesky"`
	Prometheus     PrometheusConfig `yaml:"metrics" envconfig:"metrics"`
	HTTPServer     HTTPServerConfig `yaml:"http_server" envconfig:"http_server"`
	AlertPoller    AlertPollConfig  `yaml:"alert_poller" envconfig:"alert_poller"`
	FirehoseNozzle FirehoseConfig   `yaml:"firehose" envconfig:"firehose"`
}

type contextKey struct{}

var configContextKey contextKey

func LoadAppConfig(ctx context.Context) (context.Context, error) {
	configFile := ""
	flag.StringVar(&configFile, "f", "", "Config YAML file, or '-' for stdin. If not set, use environment variables")
	flag.Parse()

	appCfg := AppConfig{}
	if configFile == "" {
		err := envconfig.Process("watchedsky", &appCfg)
		if err != nil {
			return nil, err
		}
	} else {

		yamlFile := os.Stdin
		if configFile != "-" {
			var err error
			yamlFile, err = os.Open(configFile)
			if err != nil {
				return nil, err
			}
			defer yamlFile.Close()
		}

		err := yaml.NewDecoder(yamlFile).Decode(&appCfg)
		if err != nil {
			return nil, err
		}
	}

	return context.WithValue(ctx, configContextKey, appCfg), nil
}

func GetConfig(ctx context.Context) AppConfig {
	cfg, _ := ctx.Value(configContextKey).(AppConfig)
	return cfg
}

type LogLevel string

const (
	Off   LogLevel = "off"
	Error LogLevel = "error"
	Warn  LogLevel = "warn"
	Info  LogLevel = "info"
	Debug LogLevel = "debug"
)

const SlogOff slog.Level = slog.Level(-999)

var levelMap map[LogLevel]slog.Level = map[LogLevel]slog.Level{
	Off:   SlogOff,
	Error: slog.LevelError,
	Warn:  slog.LevelWarn,
	Info:  slog.LevelInfo,
	Debug: slog.LevelDebug,
}

func (l LogLevel) SLogLevel() slog.Level {
	level, ok := levelMap[l]
	if !ok {
		return slog.LevelInfo
	}

	return level
}
