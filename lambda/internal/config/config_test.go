package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_WithAllEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalTableName := os.Getenv("TABLE_NAME")
	originalRegion := os.Getenv("AWS_REGION")
	originalLogLevel := os.Getenv("LOG_LEVEL")

	// Clean up after test
	defer func() {
		if originalTableName != "" {
			os.Setenv("TABLE_NAME", originalTableName)
		} else {
			os.Unsetenv("TABLE_NAME")
		}
		if originalRegion != "" {
			os.Setenv("AWS_REGION", originalRegion)
		} else {
			os.Unsetenv("AWS_REGION")
		}
		if originalLogLevel != "" {
			os.Setenv("LOG_LEVEL", originalLogLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	// Set test environment variables
	os.Setenv("TABLE_NAME", "test-units-table")
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("LOG_LEVEL", "DEBUG")

	config, err := New()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "test-units-table", config.TableName)
	assert.Equal(t, "us-west-2", config.Region)
	assert.Equal(t, "DEBUG", config.LogLevel)
}

func TestNew_WithMissingTableName(t *testing.T) {
	// Save original environment
	originalTableName := os.Getenv("TABLE_NAME")

	// Clean up after test
	defer func() {
		if originalTableName != "" {
			os.Setenv("TABLE_NAME", originalTableName)
		} else {
			os.Unsetenv("TABLE_NAME")
		}
	}()

	// Unset TABLE_NAME
	os.Unsetenv("TABLE_NAME")

	config, err := New()
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "TABLE_NAME environment variable is required")
}

func TestNew_WithDefaults(t *testing.T) {
	// Save original environment
	originalTableName := os.Getenv("TABLE_NAME")
	originalRegion := os.Getenv("AWS_REGION")
	originalLogLevel := os.Getenv("LOG_LEVEL")

	// Clean up after test
	defer func() {
		if originalTableName != "" {
			os.Setenv("TABLE_NAME", originalTableName)
		} else {
			os.Unsetenv("TABLE_NAME")
		}
		if originalRegion != "" {
			os.Setenv("AWS_REGION", originalRegion)
		} else {
			os.Unsetenv("AWS_REGION")
		}
		if originalLogLevel != "" {
			os.Setenv("LOG_LEVEL", originalLogLevel)
		} else {
			os.Unsetenv("LOG_LEVEL")
		}
	}()

	// Set only required environment variable
	os.Setenv("TABLE_NAME", "test-units-table")
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("LOG_LEVEL")

	config, err := New()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.Equal(t, "test-units-table", config.TableName)
	assert.Equal(t, "us-east-1", config.Region) // Default region
	assert.Equal(t, "INFO", config.LogLevel)    // Default log level
}

func TestNew_WithEmptyTableName(t *testing.T) {
	// Save original environment
	originalTableName := os.Getenv("TABLE_NAME")

	// Clean up after test
	defer func() {
		if originalTableName != "" {
			os.Setenv("TABLE_NAME", originalTableName)
		} else {
			os.Unsetenv("TABLE_NAME")
		}
	}()

	// Set empty TABLE_NAME
	os.Setenv("TABLE_NAME", "")

	config, err := New()
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "TABLE_NAME environment variable is required")
}

func TestConfig_Struct(t *testing.T) {
	config := &Config{
		TableName: "test-table",
		Region:    "eu-west-1",
		LogLevel:  "WARN",
	}

	assert.Equal(t, "test-table", config.TableName)
	assert.Equal(t, "eu-west-1", config.Region)
	assert.Equal(t, "WARN", config.LogLevel)
}
