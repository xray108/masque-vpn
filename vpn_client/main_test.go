package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		environment string
		expectError bool
	}{
		{
			name:        "Info level development",
			logLevel:    "info",
			environment: "development",
			expectError: false,
		},
		{
			name:        "Debug level production",
			logLevel:    "debug",
			environment: "production",
			expectError: false,
		},
		{
			name:        "Invalid level",
			logLevel:    "invalid",
			environment: "development",
			expectError: false, // Should default to info
		},
		{
			name:        "Empty level",
			logLevel:    "",
			environment: "development",
			expectError: false, // Should default to info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.environment != "" {
				os.Setenv("ENVIRONMENT", tt.environment)
				defer os.Unsetenv("ENVIRONMENT")
			}

			err := initLogger(tt.logLevel)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				
				// Test that logger works
				logger.Info("test message")
				
				// Cleanup
				logger.Sync()
			}
		})
	}
}

func TestLoggerConfiguration(t *testing.T) {
	// Test development configuration
	os.Setenv("ENVIRONMENT", "development")
	defer os.Unsetenv("ENVIRONMENT")
	
	err := initLogger("debug")
	require.NoError(t, err)
	require.NotNil(t, logger)
	
	// Verify logger is configured for development
	assert.True(t, logger.Core().Enabled(zapcore.DebugLevel))
	
	logger.Sync()
}

func TestLoggerProductionConfiguration(t *testing.T) {
	// Test production configuration
	os.Setenv("ENVIRONMENT", "production")
	defer os.Unsetenv("ENVIRONMENT")
	
	err := initLogger("info")
	require.NoError(t, err)
	require.NotNil(t, logger)
	
	// Verify logger is configured for production
	assert.True(t, logger.Core().Enabled(zapcore.InfoLevel))
	assert.False(t, logger.Core().Enabled(zapcore.DebugLevel))
	
	logger.Sync()
}