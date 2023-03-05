package config

import (
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
)

// GlobalConfig holds the global application configuration
type GlobalConfig struct {
	ServiceType    string `yaml:"service_type"`
	ServicesFile   string `yaml:"services_file"`
	RedisNamespace string `yaml:"redis_namespace"`
	RedisURI       string `yaml:"redis_uri"`
}

// SSLConfig holds the SSL configuration
type SSLConfig struct {
	Enabled bool   `yaml:"enabled"`
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
}

// APIConfig holds the API server configuration
type APIConfig struct {
	Addr string    `yaml:"addr"`
	SSL  SSLConfig `yaml:"ssl"`
}

// GatewayConfig holds the gateway server configuration
type GatewayConfig struct {
	Addr string    `yaml:"addr"`
	SSL  SSLConfig `yaml:"ssl"`
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

	// Check if SSL is enabled for the API server
	if apiSSL := os.Getenv("API_SSL_ENABLED"); apiSSL != "" {
		config.APIConfig.SSL.Enabled = apiSSL == "true"
	}
	if config.APIConfig.SSL.Enabled {
		if certPath := os.Getenv("API_SSL_CERT"); certPath != "" {
			config.APIConfig.SSL.Cert = certPath
		}
		if keyPath := os.Getenv("API_SSL_KEY"); keyPath != "" {
			config.APIConfig.SSL.Key = keyPath
		}
	}

	// Check if SSL is enabled for the Gateway server
	if gatewaySSL := os.Getenv("GATEWAY_SSL_ENABLED"); gatewaySSL != "" {
		config.GatewayConfig.SSL.Enabled = gatewaySSL == "true"
	}
	if config.GatewayConfig.SSL.Enabled {
		if certPath := os.Getenv("GATEWAY_SSL_CERT"); certPath != "" {
			config.GatewayConfig.SSL.Cert = certPath
		}
		if keyPath := os.Getenv("GATEWAY_SSL_KEY"); keyPath != "" {
			config.GatewayConfig.SSL.Key = keyPath
		}
	}

	return config, nil
}
