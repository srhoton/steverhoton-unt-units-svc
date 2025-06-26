package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	TableName string
	Region    string
	LogLevel  string
}

// New creates a new configuration from environment variables
func New() (*Config, error) {
	tableName := os.Getenv("TABLE_NAME")
	if tableName == "" {
		return nil, fmt.Errorf("TABLE_NAME environment variable is required")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1" // Default region
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO" // Default log level
	}

	return &Config{
		TableName: tableName,
		Region:    region,
		LogLevel:  logLevel,
	}, nil
}
