package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

// GlobalConfig holds the global application configuration
type GlobalConfig struct {
	ServiceType    string `yaml:"service_type"`
	ServicesFile   string `yaml:"services_file"`
	RedisNamespace string `yaml:"redis_namespace"`
	RedisURI       string `yaml:"redis_uri"`
}

// APIConfig holds the API server configuration
type APIConfig struct {
	Addr string `yaml:"addr"`
}

// GatewayConfig holds the gateway server configuration
type GatewayConfig struct {
	Addr string `yaml:"addr"`
}

// LoggingConfig holds the logging configuration
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// Config holds the complete application configuration
type Config struct {
	GlobalConfig  GlobalConfig  `yaml:"global"`
	APIConfig     APIConfig     `yaml:"api"`
	GatewayConfig GatewayConfig `yaml:"gateway"`
	LoggingConfig LoggingConfig `yaml:"logging"`
}

// LoadConfig loads the application configuration from a YAML file and environment variables
func LoadConfig(filename string) (*Config, error) {
	// Load the YAML configuration file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
