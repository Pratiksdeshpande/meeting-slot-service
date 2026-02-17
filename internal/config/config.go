package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	AWS      AWSConfig
	App      AppConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	// SecretARN is the ARN of the AWS Secrets Manager secret containing DB credentials
	SecretARN string
}

// AWSConfig holds AWS-specific configuration
type AWSConfig struct {
	Region string
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Environment string
	LogLevel    string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:      getEnv("DB_HOST", "localhost"),
			Port:      getEnvAsInt("DB_PORT", 3306),
			User:      getEnv("DB_USER", ""),
			Password:  getEnv("DB_PASSWORD", ""),
			Name:      getEnv("DB_NAME", "meetingslots"),
			SecretARN: getEnv("DB_SECRET_ARN", ""),
		},
		AWS: AWSConfig{
			Region: getEnv("AWS_REGION", "us-east-1"),
		},
		App: AppConfig{
			Environment: getEnv("ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
	}

	return config, nil
}

// DSN returns the MySQL database connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
		c.User, c.Password, c.Host, c.Port, c.Name,
	)
}

// Address returns the server address in host:port format
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
