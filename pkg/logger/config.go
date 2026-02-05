package logger

import "strings"

func NewFromAppConfig(level, serviceName, environment string) *Logger {
	cfg := &Config{
		Level:       parseLevel(level),
		Format:      determineFormat(environment),
		ServiceName: serviceName,
		Environment: environment,
		AddSource:   shouldAddSource(environment),
	}

	return New(cfg)
}

func parseLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func determineFormat(environment string) LogFormat {
	env := strings.ToLower(environment)
	if env == "production" || env == "prod" || env == "staging" {
		return FormatJSON
	}
	return FormatText
}

func shouldAddSource(environment string) bool {
	env := strings.ToLower(environment)
	return env == "development" || env == "dev"
}
