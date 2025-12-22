package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	ServiceName    = "orchestrated-ping"
	ServiceVersion = "1.0.0"
	ECSVersion     = "8.11.0"
)

type Config struct {
	Server      ServerConfig
	Service     ServiceConfig
	Environment string
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type ServiceConfig struct {
	Name    string
	Version string
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			ReadTimeout:     getEnvDuration("READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Service: ServiceConfig{
			Name:    ServiceName,
			Version: ServiceVersion,
		},
		Environment: getEnv("ENVIRONMENT", "development"),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	if _, err := strconv.Atoi(c.Server.Port); err != nil {
		return fmt.Errorf("invalid port number: %s", c.Server.Port)
	}

	if c.Service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	if c.Service.Version == "" {
		return fmt.Errorf("service version cannot be empty")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
