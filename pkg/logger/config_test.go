package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFromAppConfig(t *testing.T) {
	tests := []struct {
		name           string
		level          string
		serviceName    string
		environment    string
		expectedLevel  LogLevel
		expectedFormat LogFormat
		expectedSource bool
	}{
		{
			name:           "production environment",
			level:          "info",
			serviceName:    "api",
			environment:    "production",
			expectedLevel:  LevelInfo,
			expectedFormat: FormatJSON,
			expectedSource: false,
		},
		{
			name:           "development environment",
			level:          "debug",
			serviceName:    "api",
			environment:    "development",
			expectedLevel:  LevelDebug,
			expectedFormat: FormatText,
			expectedSource: true,
		},
		{
			name:           "staging environment",
			level:          "warn",
			serviceName:    "api",
			environment:    "staging",
			expectedLevel:  LevelWarn,
			expectedFormat: FormatJSON,
			expectedSource: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewFromAppConfig(tt.level, tt.serviceName, tt.environment)

			assert.Equal(t, tt.serviceName, logger.ServiceName())
			assert.Equal(t, tt.environment, logger.Environment())

			cfg := logger.GetConfig()
			assert.Equal(t, tt.expectedLevel, cfg.Level)
			assert.Equal(t, tt.expectedFormat, cfg.Format)
			assert.Equal(t, tt.expectedSource, cfg.AddSource)
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"invalid", LevelInfo},
		{"", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetermineFormat(t *testing.T) {
	tests := []struct {
		environment string
		expected    LogFormat
	}{
		{"production", FormatJSON},
		{"prod", FormatJSON},
		{"PRODUCTION", FormatJSON},
		{"staging", FormatJSON},
		{"STAGING", FormatJSON},
		{"development", FormatText},
		{"dev", FormatText},
		{"DEVELOPMENT", FormatText},
		{"test", FormatText},
		{"local", FormatText},
	}

	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			result := determineFormat(tt.environment)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldAddSource(t *testing.T) {
	tests := []struct {
		environment string
		expected    bool
	}{
		{"development", true},
		{"dev", true},
		{"DEVELOPMENT", true},
		{"DEV", true},
		{"production", false},
		{"prod", false},
		{"staging", false},
		{"test", false},
	}

	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			result := shouldAddSource(tt.environment)
			assert.Equal(t, tt.expected, result)
		})
	}
}
