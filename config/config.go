package config

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/marcelofabianov/course/pkg/retry"
	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	General    GeneralConfig
	Logger     LoggerConfig
	HTTP       HTTPConfig
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	Migrations MigrationsConfig
	JWT        JWTConfig
}

// GeneralConfig holds general application settings
type GeneralConfig struct {
	Env         string
	TZ          string
	ServiceName string
}

// LoggerConfig holds logger settings
type LoggerConfig struct {
	Level string
}

// HTTPConfig holds HTTP server settings
type HTTPConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	RequestTimeout  time.Duration
	MaxBodySize     int64
	SecurityHeaders SecurityHeadersConfig
	CORS            CORSConfig
	Compression     CompressionConfig
	RateLimit       RateLimitConfig
	CSRF            CSRFConfig
	TLS             TLSConfig
}

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	XContentTypeOptions     string
	XFrameOptions           string
	ContentSecurityPolicy   string
	ReferrerPolicy          string
	StrictTransportSecurity string
	CacheControl            string
	PermissionsPolicy       string
	XDNSPrefetchControl     string
	XDownloadOptions        string
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// CompressionConfig holds compression settings
type CompressionConfig struct {
	Enabled bool
	Level   int
}

// RateLimitConfig holds rate limiting settings
type RateLimitConfig struct {
	Enabled        bool
	Global         RateConfig
	Routes         map[string]RateConfig
	TrustedProxies []string
}

// RateConfig holds rate limit settings for a specific scope
type RateConfig struct {
	Limit  int           // requests allowed
	Window time.Duration // time window
	Burst  int           // burst allowance
}

// CSRFConfig holds CSRF protection settings
type CSRFConfig struct {
	Enabled    bool
	Secret     string
	CookieName string
	HeaderName string
	TTL        time.Duration
	Exempt     []string
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	API  APIConfig
	CORS CORSConfig
	TLS  TLSConfig
}

// APIConfig holds API server settings
type APIConfig struct {
	Host         string
	Port         int
	RateLimit    int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	MaxBodySize  int64
}

// TLSConfig holds TLS settings
type TLSConfig struct {
	Enabled     bool
	CertFile    string
	KeyFile     string
	HTTPSOnly   bool   // Force HTTPS connections only
	RedirectURL string // Custom HTTPS redirect URL (optional)
}

// DatabaseConfig holds database settings
type DatabaseConfig struct {
	Connect     DatabaseConnectConfig
	Pool        DatabasePoolConfig
	Credentials DatabaseCredentialsConfig
}

// DatabaseConnectConfig holds database connection settings
type DatabaseConnectConfig struct {
	QueryTimeout   time.Duration
	ExecTimeout    time.Duration
	BackoffMin     time.Duration
	BackoffMax     time.Duration
	BackoffFactor  int
	BackoffJitter  bool
	BackoffRetries int
}

// DatabasePoolConfig holds database connection pool settings
type DatabasePoolConfig struct {
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   time.Duration
	ConnMaxIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// DatabaseCredentialsConfig holds database credentials
type DatabaseCredentialsConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// RedisConfig holds Redis settings
type RedisConfig struct {
	Connect     RedisConnectConfig
	Pool        RedisPoolConfig
	Credentials RedisCredentialsConfig
}

// RedisConnectConfig holds Redis connection settings
type RedisConnectConfig struct {
	QueryTimeout   time.Duration
	ExecTimeout    time.Duration
	BackoffMin     time.Duration
	BackoffMax     time.Duration
	BackoffFactor  int
	BackoffJitter  bool
	BackoffRetries int
}

// RedisPoolConfig holds Redis connection pool settings
type RedisPoolConfig struct {
	MaxIdleConns   int
	MaxActiveConns int
}

// RedisCredentialsConfig holds Redis credentials
type RedisCredentialsConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// MigrationsConfig holds database migrations settings
type MigrationsConfig struct {
	Driver       string
	MigrationDir string
	FixturesDir  string
	SeedsDir     string
	DBString     string
}

// JWTConfig holds JWT authentication settings
type JWTConfig struct {
	AccessSecret    string
	RefreshSecret   string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
}

// Load reads configuration from environment variables using Viper
// .env file is the source of truth, with defaults as fallback
func Load() (*Config, error) {
	v := viper.New()

	// Set config file
	v.SetConfigFile(".env")
	v.SetConfigType("env")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		slog.Warn("No .env file found, using defaults", "error", err)
	}

	// Environment variables take precedence
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults(v)

	// Build config struct
	cfg := &Config{
		General: GeneralConfig{
			Env:         v.GetString("APP_GENERAL_ENV"),
			TZ:          v.GetString("APP_GENERAL_TZ"),
			ServiceName: v.GetString("APP_GENERAL_SERVICE_NAME"),
		},
		Logger: LoggerConfig{
			Level: v.GetString("APP_LOGGER_LEVEL"),
		},
		HTTP: HTTPConfig{
			Host:            v.GetString("APP_HTTP_HOST"),
			Port:            v.GetInt("APP_HTTP_PORT"),
			ReadTimeout:     v.GetDuration("APP_HTTP_READ_TIMEOUT"),
			WriteTimeout:    v.GetDuration("APP_HTTP_WRITE_TIMEOUT"),
			IdleTimeout:     v.GetDuration("APP_HTTP_IDLE_TIMEOUT"),
			ShutdownTimeout: v.GetDuration("APP_HTTP_SHUTDOWN_TIMEOUT"),
			RequestTimeout:  v.GetDuration("APP_HTTP_REQUEST_TIMEOUT"),
			MaxBodySize:     v.GetInt64("APP_HTTP_MAX_BODY_SIZE"),
			SecurityHeaders: SecurityHeadersConfig{
				XContentTypeOptions:     v.GetString("APP_HTTP_SECURITY_X_CONTENT_TYPE_OPTIONS"),
				XFrameOptions:           v.GetString("APP_HTTP_SECURITY_X_FRAME_OPTIONS"),
				ContentSecurityPolicy:   v.GetString("APP_HTTP_SECURITY_CONTENT_SECURITY_POLICY"),
				ReferrerPolicy:          v.GetString("APP_HTTP_SECURITY_REFERRER_POLICY"),
				StrictTransportSecurity: v.GetString("APP_HTTP_SECURITY_STRICT_TRANSPORT_SECURITY"),
				CacheControl:            v.GetString("APP_HTTP_SECURITY_CACHE_CONTROL"),
				PermissionsPolicy:       v.GetString("APP_HTTP_SECURITY_PERMISSIONS_POLICY"),
				XDNSPrefetchControl:     v.GetString("APP_HTTP_SECURITY_X_DNS_PREFETCH_CONTROL"),
				XDownloadOptions:        v.GetString("APP_HTTP_SECURITY_X_DOWNLOAD_OPTIONS"),
			},
			CORS: CORSConfig{
				Enabled:          v.GetBool("APP_HTTP_CORS_ENABLED"),
				AllowedOrigins:   parseCommaSeparated(v.GetString("APP_HTTP_CORS_ALLOWED_ORIGINS")),
				AllowedMethods:   parseCommaSeparated(v.GetString("APP_HTTP_CORS_ALLOWED_METHODS")),
				AllowedHeaders:   parseCommaSeparated(v.GetString("APP_HTTP_CORS_ALLOWED_HEADERS")),
				ExposedHeaders:   parseCommaSeparated(v.GetString("APP_HTTP_CORS_EXPOSED_HEADERS")),
				AllowCredentials: v.GetBool("APP_HTTP_CORS_ALLOW_CREDENTIALS"),
				MaxAge:           v.GetInt("APP_HTTP_CORS_MAX_AGE"),
			},
			Compression: CompressionConfig{
				Enabled: v.GetBool("APP_HTTP_COMPRESSION_ENABLED"),
				Level:   v.GetInt("APP_HTTP_COMPRESSION_LEVEL"),
			},
			TLS: TLSConfig{
				Enabled:     v.GetBool("APP_SERVER_API_TLS_ENABLED"),
				CertFile:    v.GetString("APP_SERVER_API_TLS_CERT_FILE"),
				KeyFile:     v.GetString("APP_SERVER_API_TLS_KEY_FILE"),
				HTTPSOnly:   v.GetBool("APP_SERVER_API_TLS_HTTPS_ONLY"),
				RedirectURL: v.GetString("APP_SERVER_API_TLS_REDIRECT_URL"),
			},
		},
		Server: ServerConfig{
			API: APIConfig{
				Host:         v.GetString("APP_SERVER_API_HOST"),
				Port:         v.GetInt("APP_SERVER_API_PORT"),
				RateLimit:    v.GetInt("APP_SERVER_API_RATE_LIMIT"),
				ReadTimeout:  v.GetDuration("APP_SERVER_API_READ_TIMEOUT"),
				WriteTimeout: v.GetDuration("APP_SERVER_API_WRITE_TIMEOUT"),
				IdleTimeout:  v.GetDuration("APP_SERVER_API_IDLE_TIMEOUT"),
				MaxBodySize:  v.GetInt64("APP_SERVER_API_MAXBODYSIZE"),
			},
			CORS: CORSConfig{
				AllowedOrigins:   parseCommaSeparated(v.GetString("APP_SERVER_CORS_ALLOWEDORIGINS")),
				AllowedMethods:   parseCommaSeparated(v.GetString("APP_SERVER_CORS_ALLOWEDMETHODS")),
				AllowedHeaders:   parseCommaSeparated(v.GetString("APP_SERVER_CORS_ALLOWEDHEADERS")),
				ExposedHeaders:   parseCommaSeparated(v.GetString("APP_SERVER_CORS_EXPOSEDHEADERS")),
				AllowCredentials: v.GetBool("APP_SERVER_CORS_ALLOWCREDENTIALS"),
			},
			TLS: TLSConfig{
				Enabled:     v.GetBool("APP_SERVER_API_TLS_ENABLED"),
				CertFile:    v.GetString("APP_SERVER_API_TLS_CERT_FILE"),
				KeyFile:     v.GetString("APP_SERVER_API_TLS_KEY_FILE"),
				HTTPSOnly:   v.GetBool("APP_SERVER_API_TLS_HTTPS_ONLY"),
				RedirectURL: v.GetString("APP_SERVER_API_TLS_REDIRECT_URL"),
			},
		},
		Database: DatabaseConfig{
			Connect: DatabaseConnectConfig{
				QueryTimeout:   v.GetDuration("APP_DB_CONNECT_QUERY_TIMEOUT"),
				ExecTimeout:    v.GetDuration("APP_DB_CONNECT_EXEC_TIMEOUT"),
				BackoffMin:     v.GetDuration("APP_DB_CONNECT_BACKOFF_MIN"),
				BackoffMax:     v.GetDuration("APP_DB_CONNECT_BACKOFF_MAX"),
				BackoffFactor:  v.GetInt("APP_DB_CONNECT_BACKOFF_FACTOR"),
				BackoffJitter:  v.GetBool("APP_DB_CONNECT_BACKOFF_JITTER"),
				BackoffRetries: v.GetInt("APP_DB_CONNECT_BACKOFF_RETRIES"),
			},
			Pool: DatabasePoolConfig{
				MaxOpenConns:      v.GetInt("APP_DB_POOL_MAX_OPEN_CONNS"),
				MaxIdleConns:      v.GetInt("APP_DB_POOL_MAX_IDLE_CONNS"),
				ConnMaxLifetime:   v.GetDuration("APP_DB_POOL_CONN_MAX_LIFETIME"),
				ConnMaxIdleTime:   v.GetDuration("APP_DB_POOL_CONN_MAX_IDLE_TIME"),
				HealthCheckPeriod: v.GetDuration("APP_DB_POOL_HEALTH_CHECK_PERIOD"),
			},
			Credentials: DatabaseCredentialsConfig{
				Host:     v.GetString("APP_DB_HOST"),
				Port:     v.GetInt("APP_DB_PORT"),
				Name:     v.GetString("APP_DB_NAME"),
				User:     v.GetString("APP_DB_USER"),
				Password: v.GetString("APP_DB_PASSWORD"),
				SSLMode:  v.GetString("APP_DB_SSL_MODE"),
			},
		},
		Redis: RedisConfig{
			Connect: RedisConnectConfig{
				QueryTimeout:   v.GetDuration("APP_REDIS_CONNECT_QUERY_TIMEOUT"),
				ExecTimeout:    v.GetDuration("APP_REDIS_CONNECT_EXEC_TIMEOUT"),
				BackoffMin:     v.GetDuration("APP_REDIS_CONNECT_BACKOFF_MIN"),
				BackoffMax:     v.GetDuration("APP_REDIS_CONNECT_BACKOFF_MAX"),
				BackoffFactor:  v.GetInt("APP_REDIS_CONNECT_BACKOFF_FACTOR"),
				BackoffJitter:  v.GetBool("APP_REDIS_CONNECT_BACKOFF_JITTER"),
				BackoffRetries: v.GetInt("APP_REDIS_CONNECT_BACKOFF_RETRIES"),
			},
			Pool: RedisPoolConfig{
				MaxIdleConns:   v.GetInt("APP_REDIS_POOL_MAX_IDLE_CONNS"),
				MaxActiveConns: v.GetInt("APP_REDIS_POOL_MAX_ACTIVE_CONNS"),
			},
			Credentials: RedisCredentialsConfig{
				Host:     v.GetString("APP_REDIS_HOST"),
				Port:     v.GetInt("APP_REDIS_PORT"),
				Password: v.GetString("APP_REDIS_PASSWORD"),
				DB:       v.GetInt("APP_REDIS_DB"),
			},
		},
		Migrations: MigrationsConfig{
			Driver:       v.GetString("GOOSE_DRIVER"),
			MigrationDir: v.GetString("GOOSE_MIGRATION_DIR"),
			FixturesDir:  v.GetString("GOOSE_FIXTURES_DIR"),
			SeedsDir:     v.GetString("GOOSE_SEEDS_DIR"),
			DBString:     v.GetString("GOOSE_DBSTRING"),
		},
		JWT: JWTConfig{
			AccessSecret:    v.GetString("APP_JWT_ACCESS_SECRET"),
			RefreshSecret:   v.GetString("APP_JWT_REFRESH_SECRET"),
			AccessTokenTTL:  v.GetDuration("APP_JWT_ACCESS_TOKEN_TTL"),
			RefreshTokenTTL: v.GetDuration("APP_JWT_REFRESH_TOKEN_TTL"),
			Issuer:          v.GetString("APP_JWT_ISSUER"),
		},
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// setDefaults configures default values for all settings
func setDefaults(v *viper.Viper) {
	// General defaults
	v.SetDefault("APP_GENERAL_ENV", "development")
	v.SetDefault("APP_GENERAL_TZ", "UTC")
	v.SetDefault("APP_GENERAL_SERVICE_NAME", "course-api")

	// Logger defaults
	v.SetDefault("APP_LOGGER_LEVEL", "info")

	// Server API defaults
	v.SetDefault("APP_SERVER_API_HOST", "0.0.0.0")
	v.SetDefault("APP_SERVER_API_PORT", 8080)
	v.SetDefault("APP_SERVER_API_RATE_LIMIT", 100)
	v.SetDefault("APP_SERVER_API_READ_TIMEOUT", "5s")
	v.SetDefault("APP_SERVER_API_WRITE_TIMEOUT", "10s")
	v.SetDefault("APP_SERVER_API_IDLE_TIMEOUT", "120s")
	v.SetDefault("APP_SERVER_API_MAXBODYSIZE", 1048576) // 1MB

	// CORS defaults
	v.SetDefault("APP_SERVER_CORS_ALLOWEDORIGINS", "http://localhost:3000")
	v.SetDefault("APP_SERVER_CORS_ALLOWEDMETHODS", "GET,POST,PUT,DELETE,PATCH")
	v.SetDefault("APP_SERVER_CORS_ALLOWEDHEADERS", "Accept,Authorization,Content-Type,X-CSRF-Token")
	v.SetDefault("APP_SERVER_CORS_EXPOSEDHEADERS", "Link")
	v.SetDefault("APP_SERVER_CORS_ALLOWCREDENTIALS", true)

	// TLS defaults
	v.SetDefault("APP_SERVER_API_TLS_ENABLED", false)
	v.SetDefault("APP_SERVER_API_TLS_CERT_FILE", "./certs/cert.pem")
	v.SetDefault("APP_SERVER_API_TLS_KEY_FILE", "./certs/key.pem")
	v.SetDefault("APP_SERVER_API_TLS_HTTPS_ONLY", true)
	v.SetDefault("APP_SERVER_API_TLS_REDIRECT_URL", "")

	// HTTP Server defaults
	v.SetDefault("APP_HTTP_HOST", "0.0.0.0")
	v.SetDefault("APP_HTTP_PORT", 8080)
	v.SetDefault("APP_HTTP_READ_TIMEOUT", "10s")
	v.SetDefault("APP_HTTP_WRITE_TIMEOUT", "10s")
	v.SetDefault("APP_HTTP_IDLE_TIMEOUT", "120s")
	v.SetDefault("APP_HTTP_SHUTDOWN_TIMEOUT", "30s")
	v.SetDefault("APP_HTTP_REQUEST_TIMEOUT", "30s")
	v.SetDefault("APP_HTTP_MAX_BODY_SIZE", 1048576) // 1MB

	// Security Headers defaults
	v.SetDefault("APP_HTTP_SECURITY_X_CONTENT_TYPE_OPTIONS", "nosniff")
	v.SetDefault("APP_HTTP_SECURITY_X_FRAME_OPTIONS", "DENY")
	v.SetDefault("APP_HTTP_SECURITY_CONTENT_SECURITY_POLICY", "default-src 'none'")
	v.SetDefault("APP_HTTP_SECURITY_REFERRER_POLICY", "no-referrer")
	v.SetDefault("APP_HTTP_SECURITY_STRICT_TRANSPORT_SECURITY", "max-age=31536000; includeSubDomains")
	v.SetDefault("APP_HTTP_SECURITY_CACHE_CONTROL", "no-store, no-cache, must-revalidate")
	v.SetDefault("APP_HTTP_SECURITY_PERMISSIONS_POLICY", "camera=(), microphone=(), geolocation=()")
	v.SetDefault("APP_HTTP_SECURITY_X_DNS_PREFETCH_CONTROL", "off")
	v.SetDefault("APP_HTTP_SECURITY_X_DOWNLOAD_OPTIONS", "noopen")

	// CORS defaults
	v.SetDefault("APP_HTTP_CORS_ENABLED", false)
	v.SetDefault("APP_HTTP_CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	v.SetDefault("APP_HTTP_CORS_ALLOWED_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	v.SetDefault("APP_HTTP_CORS_ALLOWED_HEADERS", "Accept,Content-Type,Authorization")
	v.SetDefault("APP_HTTP_CORS_EXPOSED_HEADERS", "X-Request-ID")
	v.SetDefault("APP_HTTP_CORS_ALLOW_CREDENTIALS", false)
	v.SetDefault("APP_HTTP_CORS_MAX_AGE", 300)

	// Compression defaults
	v.SetDefault("APP_HTTP_COMPRESSION_ENABLED", true)
	v.SetDefault("APP_HTTP_COMPRESSION_LEVEL", 5)

	// Database connect defaults
	v.SetDefault("APP_DB_CONNECT_QUERY_TIMEOUT", "5s")
	v.SetDefault("APP_DB_CONNECT_EXEC_TIMEOUT", "5s")
	v.SetDefault("APP_DB_CONNECT_BACKOFF_MIN", "200ms")
	v.SetDefault("APP_DB_CONNECT_BACKOFF_MAX", "15s")
	v.SetDefault("APP_DB_CONNECT_BACKOFF_FACTOR", 2)
	v.SetDefault("APP_DB_CONNECT_BACKOFF_JITTER", true)
	v.SetDefault("APP_DB_CONNECT_BACKOFF_RETRIES", 7)

	// Database pool defaults
	v.SetDefault("APP_DB_POOL_MAX_OPEN_CONNS", 20)
	v.SetDefault("APP_DB_POOL_MAX_IDLE_CONNS", 10)
	v.SetDefault("APP_DB_POOL_CONN_MAX_LIFETIME", "15m")
	v.SetDefault("APP_DB_POOL_CONN_MAX_IDLE_TIME", "5m")
	v.SetDefault("APP_DB_POOL_HEALTH_CHECK_PERIOD", "1m")

	// Database credentials defaults
	v.SetDefault("APP_DB_HOST", "localhost")
	v.SetDefault("APP_DB_PORT", 5432)
	v.SetDefault("APP_DB_NAME", "course")
	v.SetDefault("APP_DB_USER", "course")
	v.SetDefault("APP_DB_PASSWORD", "")
	v.SetDefault("APP_DB_SSL_MODE", "disable")

	// Redis connect defaults
	v.SetDefault("APP_REDIS_CONNECT_QUERY_TIMEOUT", "2s")
	v.SetDefault("APP_REDIS_CONNECT_EXEC_TIMEOUT", "2s")
	v.SetDefault("APP_REDIS_CONNECT_BACKOFF_MIN", "200ms")
	v.SetDefault("APP_REDIS_CONNECT_BACKOFF_MAX", "15s")
	v.SetDefault("APP_REDIS_CONNECT_BACKOFF_FACTOR", 2)
	v.SetDefault("APP_REDIS_CONNECT_BACKOFF_JITTER", true)
	v.SetDefault("APP_REDIS_CONNECT_BACKOFF_RETRIES", 7)

	// Redis pool defaults
	v.SetDefault("APP_REDIS_POOL_MAX_IDLE_CONNS", 10)
	v.SetDefault("APP_REDIS_POOL_MAX_ACTIVE_CONNS", 20)

	// Redis credentials defaults
	v.SetDefault("APP_REDIS_HOST", "localhost")
	v.SetDefault("APP_REDIS_PORT", 6379)
	v.SetDefault("APP_REDIS_PASSWORD", "")
	v.SetDefault("APP_REDIS_DB", 0)

	// Migrations defaults
	v.SetDefault("GOOSE_DRIVER", "postgres")
	v.SetDefault("GOOSE_MIGRATION_DIR", "./db/migrations")
	v.SetDefault("GOOSE_FIXTURES_DIR", "./db/fixtures")
	v.SetDefault("GOOSE_SEEDS_DIR", "./db/seeds")

	// JWT defaults
	v.SetDefault("APP_JWT_ACCESS_SECRET", "dev-access-secret-min-32-bytes-change-in-production!")
	v.SetDefault("APP_JWT_REFRESH_SECRET", "dev-refresh-secret-min-32-bytes-change-in-production!")
	v.SetDefault("APP_JWT_ACCESS_TOKEN_TTL", "15m")
	v.SetDefault("APP_JWT_REFRESH_TOKEN_TTL", "168h") // 7 days
	v.SetDefault("APP_JWT_ISSUER", "course-api")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
		"test":        true,
	}
	if !validEnvs[c.General.Env] {
		return fmt.Errorf("invalid environment: %s (must be development, staging, production, or test)", c.General.Env)
	}

	// Validate logger level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[strings.ToLower(c.Logger.Level)] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Logger.Level)
	}

	// Validate server port
	if c.Server.API.Port < 1 || c.Server.API.Port > 65535 {
		return fmt.Errorf("invalid API port: %d (must be between 1 and 65535)", c.Server.API.Port)
	}

	// Validate timeouts
	if c.Server.API.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	if c.Server.API.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	// Validate TLS configuration
	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" {
			return fmt.Errorf("TLS cert file is required when TLS is enabled")
		}
		if c.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file is required when TLS is enabled")
		}
	}

	// Validate database configuration
	if c.Database.Credentials.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Credentials.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Database.Credentials.User == "" {
		return fmt.Errorf("database user is required")
	}

	// Validate database pool
	if c.Database.Pool.MaxOpenConns < 1 {
		return fmt.Errorf("max open connections must be at least 1")
	}
	if c.Database.Pool.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections must be non-negative")
	}
	if c.Database.Pool.MaxIdleConns > c.Database.Pool.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot exceed max open connections")
	}

	// Validate Redis configuration
	if c.Redis.Credentials.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if c.Redis.Credentials.Port < 1 || c.Redis.Credentials.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", c.Redis.Credentials.Port)
	}

	// Validate JWT configuration
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT access secret is required")
	}
	if len(c.JWT.AccessSecret) < 32 {
		return fmt.Errorf("JWT access secret must be at least 32 bytes")
	}
	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT refresh secret is required")
	}
	if len(c.JWT.RefreshSecret) < 32 {
		return fmt.Errorf("JWT refresh secret must be at least 32 bytes")
	}
	if c.JWT.AccessTokenTTL <= 0 {
		return fmt.Errorf("JWT access token TTL must be positive")
	}
	if c.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("JWT refresh token TTL must be positive")
	}
	if c.JWT.Issuer == "" {
		return fmt.Errorf("JWT issuer is required")
	}

	return nil
}

// GetDatabaseDSN returns the PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Database.Credentials.Host,
		c.Database.Credentials.Port,
		c.Database.Credentials.Name,
		c.Database.Credentials.User,
		c.Database.Credentials.Password,
		c.Database.Credentials.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Credentials.Host, c.Redis.Credentials.Port)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.General.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.General.Env == "production"
}

// GetDatabaseRetryConfig returns a retry configuration for database operations
func (c *Config) GetDatabaseRetryConfig() *retry.Config {
	return &retry.Config{
		MaxAttempts: c.Database.Connect.BackoffRetries,
		Strategy: retry.NewExponentialBackoff(retry.ExponentialBackoffConfig{
			Min:    c.Database.Connect.BackoffMin,
			Max:    c.Database.Connect.BackoffMax,
			Factor: float64(c.Database.Connect.BackoffFactor),
			Jitter: c.Database.Connect.BackoffJitter,
		}),
		OnRetry: func(attempt int, err error) {
			slog.Warn("Database operation retry",
				"attempt", attempt+1,
				"max_attempts", c.Database.Connect.BackoffRetries,
				"error", err.Error(),
			)
		},
	}
}

// GetRedisRetryConfig returns a retry configuration for Redis operations
func (c *Config) GetRedisRetryConfig() *retry.Config {
	return &retry.Config{
		MaxAttempts: c.Redis.Connect.BackoffRetries,
		Strategy: retry.NewExponentialBackoff(retry.ExponentialBackoffConfig{
			Min:    c.Redis.Connect.BackoffMin,
			Max:    c.Redis.Connect.BackoffMax,
			Factor: float64(c.Redis.Connect.BackoffFactor),
			Jitter: c.Redis.Connect.BackoffJitter,
		}),
		OnRetry: func(attempt int, err error) {
			slog.Warn("Redis operation retry",
				"attempt", attempt+1,
				"max_attempts", c.Redis.Connect.BackoffRetries,
				"error", err.Error(),
			)
		},
	}
}

// parseCommaSeparated splits a comma-separated string into a slice
func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
