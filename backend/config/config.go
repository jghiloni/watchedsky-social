package config

import (
	"flag"
	"os"
	"time"

	"github.com/jghiloni/watchedsky-social/backend/logging"
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

type RabbitMQConfig struct {
	Host     string `yaml:"host" envconfig:"host"`
	Username string `yaml:"username" envconfig:"username"`
	Password string `yaml:"password" envconfig:"password"`
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

type DBLoaderConfig struct {
	Enabled bool `yaml:"enabled" envconfig:"enabled"`
}

type SkeeterConfig struct {
	Enabled bool `yaml:"enabled" envconfig:"enabled"`
}

type AppConfig struct {
	LogLevel    logging.LogLevel `yaml:"log_level" envconfig:"log_level"`
	MongoDB     DatabaseConfig   `yaml:"database" envconfig:"database"`
	RabbitMQ    RabbitMQConfig   `yaml:"rabbitmq" envconfig:"rabbitmq"`
	Bluesky     BlueskyConfig    `yaml:"bluesky" envconfig:"bluesky"`
	Prometheus  PrometheusConfig `yaml:"metrics" envconfig:"metrics"`
	HTTPServer  HTTPServerConfig `yaml:"http_server" envconfig:"http_server"`
	AlertPoller AlertPollConfig  `yaml:"alert_poller" envconfig:"alert_poller"`
	DBLoader    DBLoaderConfig   `yaml:"db_loader" envconfig:"db_loader"`
	Skeeter     SkeeterConfig    `yaml:"skeeter" envconfig:"skeeter"`
}

func LoadAppConfig() (*AppConfig, error) {
	configFile := ""
	flag.StringVar(&configFile, "f", "", "Config YAML file, or '-' for stdin. If not set, use environment variables")
	flag.Parse()

	appCfg := AppConfig{}
	if configFile == "" {
		err := envconfig.Process("watchedsky", &appCfg)
		if err != nil {
			return nil, err
		}

		return &appCfg, nil
	}

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

	return &appCfg, nil
}
