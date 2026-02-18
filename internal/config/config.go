package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	AWS      AWSConfig
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	// SecretARN is the ARN of the AWS Secrets Manager secret containing DB
	// credentials. When set, Host/Port/User/Password/Name are ignored.
	SecretARN string
}

// AWSConfig holds AWS-specific configuration.
type AWSConfig struct {
	Region string
}

// Load reads configuration from environment variables and validates required
// fields. Returns an error if any required value is missing or malformed.
func Load() (*Config, error) {
	serverPort, err := getEnvAsInt("SERVER_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	dbPort, err := getEnvAsInt("DB_PORT", 3306)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: serverPort,
		},
		Database: DatabaseConfig{
			Host:      getEnv("DB_HOST", "localhost"),
			Port:      dbPort,
			User:      getEnv("DB_USER", ""),
			Password:  getEnv("DB_PASSWORD", ""),
			Name:      getEnv("DB_NAME", "meetingslots"),
			SecretARN: getEnv("DB_SECRET_ARN", ""),
		},
		AWS: AWSConfig{
			Region: getEnv("AWS_REGION", "us-east-1"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate checks that all required fields are present.
func (c *Config) validate() error {
	// When SecretARN is set, credentials come from Secrets Manager — no local
	// user/password required.
	if c.Database.SecretARN == "" && c.Database.User == "" {
		return fmt.Errorf("DB_USER must be set (or provide DB_SECRET_ARN for AWS Secrets Manager)")
	}
	return nil
}

// DSN returns the MySQL data-source name for direct (non-Secrets Manager) use.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
		c.User, c.Password, c.Host, c.Port, c.Name,
	)
}

// Address returns the server listen address in host:port format.
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// getEnv returns the value of key, or defaultValue when the variable is unset
// or empty.
func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

// getEnvAsInt parses key as an integer. Returns defaultValue when the variable
// is unset, and an error when it is set but not a valid integer.
func getEnvAsInt(key string, defaultValue int) (int, error) {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("environment variable %s=%q is not a valid integer", key, s)
	}
	return v, nil
}
