package config_test

import (
	"os"
	"testing"

	"github.com/marcelofabianov/course/config"
)

func TestLoad(t *testing.T) {
	// Save original env
	originalEnv := os.Getenv("APP_GENERAL_ENV")
	defer func() {
		if originalEnv != "" {
			os.Setenv("APP_GENERAL_ENV", originalEnv)
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		check   func(*testing.T, *config.Config)
	}{
		{
			name: "load with defaults",
			envVars: map[string]string{
				"APP_DB_USER":     "testuser",
				"APP_DB_PASSWORD": "testpass",
				"APP_DB_NAME":     "testdb",
				"APP_REDIS_HOST":  "localhost",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.General.Env != "development" {
					t.Errorf("Expected env 'development', got '%s'", cfg.General.Env)
				}
				if cfg.Server.API.Port != 8080 {
					t.Errorf("Expected port 8080, got %d", cfg.Server.API.Port)
				}
			},
		},
		{
			name: "load with custom values",
			envVars: map[string]string{
				"APP_GENERAL_ENV":     "production",
				"APP_SERVER_API_PORT": "9000",
				"APP_LOGGER_LEVEL":    "warn",
				"APP_DB_HOST":         "prod-db.example.com",
				"APP_DB_PORT":         "5432",
				"APP_DB_NAME":         "proddb",
				"APP_DB_USER":         "produser",
				"APP_DB_PASSWORD":     "prodpass",
				"APP_REDIS_HOST":      "redis.example.com",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *config.Config) {
				if cfg.General.Env != "production" {
					t.Errorf("Expected env 'production', got '%s'", cfg.General.Env)
				}
				if cfg.Server.API.Port != 9000 {
					t.Errorf("Expected port 9000, got %d", cfg.Server.API.Port)
				}
				if cfg.Logger.Level != "warn" {
					t.Errorf("Expected log level 'warn', got '%s'", cfg.Logger.Level)
				}
			},
		},
		{
			name: "invalid environment",
			envVars: map[string]string{
				"APP_GENERAL_ENV": "invalid",
				"APP_DB_USER":     "testuser",
				"APP_DB_NAME":     "testdb",
				"APP_REDIS_HOST":  "localhost",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			envVars: map[string]string{
				"APP_SERVER_API_PORT": "99999",
				"APP_DB_USER":         "testuser",
				"APP_DB_NAME":         "testdb",
				"APP_REDIS_HOST":      "localhost",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearTestEnv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load configuration
			cfg, err := config.Load()

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Run additional checks if provided
			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}

			// Cleanup
			clearTestEnv()
		})
	}
}

func TestGetDatabaseDSN(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Credentials: config.DatabaseCredentialsConfig{
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				User:     "testuser",
				Password: "testpass",
				SSLMode:  "disable",
			},
		},
	}

	expected := "host=localhost port=5432 dbname=testdb user=testuser password=testpass sslmode=disable"
	got := cfg.GetDatabaseDSN()

	if got != expected {
		t.Errorf("GetDatabaseDSN() = %v, want %v", got, expected)
	}
}

func TestGetRedisAddr(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{
			Credentials: config.RedisCredentialsConfig{
				Host: "redis.example.com",
				Port: 6379,
			},
		},
	}

	expected := "redis.example.com:6379"
	got := cfg.GetRedisAddr()

	if got != expected {
		t.Errorf("GetRedisAddr() = %v, want %v", got, expected)
	}
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"development", "development", true},
		{"production", "production", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				General: config.GeneralConfig{Env: tt.env},
			}
			if got := cfg.IsDevelopment(); got != tt.want {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"production", "production", true},
		{"development", "development", false},
		{"staging", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				General: config.GeneralConfig{Env: tt.env},
			}
			if got := cfg.IsProduction(); got != tt.want {
				t.Errorf("IsProduction() = %v, want %v", got, tt.want)
			}
		})
	}
}

// clearTestEnv clears test environment variables
func clearTestEnv() {
	envVars := []string{
		"APP_GENERAL_ENV",
		"APP_GENERAL_TZ",
		"APP_GENERAL_SERVICE_NAME",
		"APP_LOGGER_LEVEL",
		"APP_SERVER_API_HOST",
		"APP_SERVER_API_PORT",
		"APP_SERVER_API_RATE_LIMIT",
		"APP_DB_HOST",
		"APP_DB_PORT",
		"APP_DB_NAME",
		"APP_DB_USER",
		"APP_DB_PASSWORD",
		"APP_DB_SSL_MODE",
		"APP_REDIS_HOST",
		"APP_REDIS_PORT",
		"APP_REDIS_PASSWORD",
		"APP_REDIS_DB",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}
}

func TestGetDatabaseRetryConfig(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	retryConfig := cfg.GetDatabaseRetryConfig()
	if retryConfig == nil {
		t.Fatal("Expected retry config, got nil")
	}

	if retryConfig.Strategy == nil {
		t.Error("Expected strategy to be set")
	}

	if retryConfig.MaxAttempts <= 0 {
		t.Errorf("Expected MaxAttempts > 0, got %d", retryConfig.MaxAttempts)
	}

	if retryConfig.OnRetry == nil {
		t.Error("Expected OnRetry callback to be set")
	}
}

func TestGetRedisRetryConfig(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	retryConfig := cfg.GetRedisRetryConfig()
	if retryConfig == nil {
		t.Fatal("Expected retry config, got nil")
	}

	if retryConfig.Strategy == nil {
		t.Error("Expected strategy to be set")
	}

	if retryConfig.MaxAttempts <= 0 {
		t.Errorf("Expected MaxAttempts > 0, got %d", retryConfig.MaxAttempts)
	}

	if retryConfig.OnRetry == nil {
		t.Error("Expected OnRetry callback to be set")
	}
}
