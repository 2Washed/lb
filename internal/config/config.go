package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const CONFIG_FILE_ENV = "LB_CONFIG"
const DEFAULT_CONFIG_FILE = "config.json"

type Configuration struct {
	Port                uint16                 `json:"port"`
	HealthCheckInterval Duration               `json:"healthCheckInterval"`
	MaxRetries          int                    `json:"maxRetries"`
	Servers             []*ServerConfiguration `json:"servers"`
	RateLimiter         *RateLimiterConfig     `json:"rateLimiter"`
	BalancingAlgorithm  BalancingAlgorithm     `json:"balancer"` //this will default to 0 => round robin
}

func readConfigurationFile() []byte {
	filename, ok := os.LookupEnv(CONFIG_FILE_ENV)
	if !ok {
		filename = DEFAULT_CONFIG_FILE
	}

	configFile, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("[ERROR] Could not read configuration file: %v, error: %v", filename, err)
	}

	return configFile
}

func GetConfiguration() *Configuration {
	configFileContent := readConfigurationFile()
	var configuration Configuration
	err := json.Unmarshal(configFileContent, &configuration)
	if err != nil {
		log.Fatalf("[ERROR] Could not parse configuration file, error: %v", err)
	}

	if err := configuration.validate(); err != nil {
		log.Fatalf("[ERROR] Invalid configuration: %v\n", err)
	}

	return &configuration
}

func (c *Configuration) validate() error {
	if len(c.Servers) == 0 {
		return fmt.Errorf("no servers defined in config")
	}

	if c.Port == 0 {
		return fmt.Errorf("port is required")
	}

	if c.MaxRetries < 0 {
		return fmt.Errorf("maxRetries cannot be negative")
	}

	return nil
}
