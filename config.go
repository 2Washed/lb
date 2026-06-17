package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

const CONFIG_FILE_ENV = "LB_CONFIG"
const DEFAULT_CONFIG_FILE = "config.json"

type ServerConfiguration struct {
	Url    string `json:"url"`
	Weight int    `json:"weight"`
}

type RateLimiterConfig struct {
	Rate         int      `json:"rate"`
	BurstSeconds int      `json:"burstSeconds"`
	Expiry       Duration `json:"expiry"`
}

type Configuration struct {
	Port                uint16                 `json:"port"`
	HealthCheckInterval Duration               `json:"healthCheckInterval"`
	MaxRetries          int                    `json:"maxRetries"`
	Servers             []*ServerConfiguration `json:"servers"`
	RateLimiter         *RateLimiterConfig     `json:"rateLimiter"`
	BalancingAlgorithm  BalancingAlgorithm     `json:"balancer"` //this will default to 0 => round robin
}

type Duration struct {
	time.Duration
}

type BalancingAlgorithm int

const (
	RoundRobinAlgo BalancingAlgorithm = iota //TODO remove 'Algo from the name after moving to separate packages'
	LeastConnectionsAlgo
)

var stringToAlgo = map[string]BalancingAlgorithm{
	"round-robin":       RoundRobinAlgo,
	"least-connections": LeastConnectionsAlgo,
}

var algoToString = map[BalancingAlgorithm]string{
	RoundRobinAlgo:       "round-robin",
	LeastConnectionsAlgo: "least-connections",
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = duration
	return nil
}

func (algo *BalancingAlgorithm) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	algorithm, ok := stringToAlgo[s]
	if !ok {
		return fmt.Errorf("invalid balancing algorithm: %s", s)
	}

	*algo = algorithm
	return nil
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

func getConfiguration() *Configuration {
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
