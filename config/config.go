package config

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is kafka-firehose-nozzle configuration.
type Config struct {
	ScrapeEndpoints []ScrapeEndpoint
	Elasticsearch   Elasticsearch
	MongoDB         MongoDB
}

// Elasticsearch contains the Cluster Configuration for the
// ES Endpoint
type Elasticsearch struct {
	Hosts []string `env:"ES_HOST"`
	Port  int      `env:"ES_PORT"`
	HTTPS bool     `env:"ES_USE_HTTPS"`
	// Username is the username which can has scope of `doppler.firehose`.
	Username string `env:"ES_USERNAME"`
	Password string `env:"ES_PASSWORD"`
}

// ScrapeEndpoint contains the Configuration for all relevant Scrape
// Endpoints of the Prometheus Agents
type ScrapeEndpoint struct {
	Type        string `env:"SE_TYPE"`
	Port        int    `env:"SE_PORT"`
	Interval    int    `env:"SE_INTERVAL"`
	IncludeNode bool   `env:"SE_INLCUDE_NODE"`
}

// MongoDB contains the Cluster Configuration for the
// MongoDB Endpoint
type MongoDB struct {
	Hosts []string `env:"MDB_HOST"`
	Port  int      `env:"MDB_PORT"`
	// Username is the username which can has scope of `doppler.firehose`.
	Username   string `env:"MDB_USERNAME"`
	Password   string `env:"MDB_PASSWORD"`
	Database   string `env:"MDB_DATABASE"`
	Collection string `env:"MDB_COLLECTION"`
}

// LoadConfig reads configuration file
func LoadConfig(path string) (*Config, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	config := new(Config)
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	return config, nil
}
