// Package database fornece um wrapper para conexão com PostgreSQL usando pgx driver.
//
// O pacote implementa:
// - Retry automático em falhas de conexão
// - Connection pooling com configurações otimizadas
// - Health checks (manual e background)
// - Tratamento de erros estruturado com fault
// - Timeouts configuráveis para todas as operações
package database

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/marcelofabianov/course/config"
	"github.com/marcelofabianov/course/pkg/retry"
	"github.com/marcelofabianov/fault"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

var (
	// ErrConnectionFailed is returned when database connection fails after all retries
	ErrConnectionFailed = fault.New(
		"database connection failed after retries",
		fault.WithCode(fault.InfraError),
	)

	// ErrInvalidConfig is returned when database configuration is invalid
	ErrInvalidConfig = fault.New(
		"invalid database configuration",
		fault.WithCode(fault.Invalid),
	)

	// ErrAlreadyConnected is returned when attempting to connect an already connected database
	ErrAlreadyConnected = fault.New(
		"database already connected",
		fault.WithCode(fault.Conflict),
	)

	// ErrNotConnected is returned when attempting operations on a non-connected database
	ErrNotConnected = fault.New(
		"database not connected",
		fault.WithCode(fault.NotFound),
	)

	// ErrOpenFailed is returned when sql.Open fails
	ErrOpenFailed = fault.New(
		"failed to open database connection",
		fault.WithCode(fault.InfraError),
	)

	// ErrPingFailed is returned when database ping fails
	ErrPingFailed = fault.New(
		"failed to ping database",
		fault.WithCode(fault.InfraError),
	)

	// ErrCloseFailed is returned when database close fails
	ErrCloseFailed = fault.New(
		"failed to close database connection",
		fault.WithCode(fault.Internal),
	)

	// ErrExecFailed is returned when query execution fails
	ErrExecFailed = fault.New(
		"failed to execute query",
		fault.WithCode(fault.Internal),
	)

	// ErrQueryFailed is returned when query fails
	ErrQueryFailed = fault.New(
		"failed to execute query",
		fault.WithCode(fault.Internal),
	)

	// ErrTransactionFailed is returned when transaction fails to begin
	ErrTransactionFailed = fault.New(
		"failed to begin transaction",
		fault.WithCode(fault.Internal),
	)
)

// DB wraps sql.DB with additional functionality
type DB struct {
	conn   *sql.DB
	config *config.Config
	logger *slog.Logger
}

// New creates a new database instance with the given configuration
func New(cfg *config.Config) (*DB, error) {
	if cfg == nil {
		return nil, ErrInvalidConfig
	}

	return &DB{
		config: cfg,
		logger: slog.Default(),
	}, nil
}

// SetLogger sets a custom logger for the database instance
func (db *DB) SetLogger(logger *slog.Logger) {
	if logger != nil {
		db.logger = logger
	}
}

// Connect establishes a connection to the database with retry logic
func (db *DB) Connect(ctx context.Context) error {
	if db.conn != nil {
		return ErrAlreadyConnected
	}

	db.logger.Info("Connecting to database",
		"host", db.config.Database.Credentials.Host,
		"database", db.config.Database.Credentials.Name,
		"max_retries", db.config.Database.Connect.BackoffRetries,
	)

	retryConfig := db.config.GetDatabaseRetryConfig()
	retryConfig.Logger = db.logger

	err := retry.Do(ctx, retryConfig, func(ctx context.Context) error {
		return db.connect(ctx)
	})
	if err != nil {
		db.logger.Error("Failed to connect to database",
			"host", db.config.Database.Credentials.Host,
			"database", db.config.Database.Credentials.Name,
			"error", err.Error(),
		)

		if fault.IsCode(err, fault.Invalid) {
			return fault.Wrap(ErrConnectionFailed, "connection failed after all retries",
				fault.WithWrappedErr(err),
				fault.WithContext("host", db.config.Database.Credentials.Host),
				fault.WithContext("database", db.config.Database.Credentials.Name),
				fault.WithContext("retries", db.config.Database.Connect.BackoffRetries),
			)
		}
		return fault.Wrap(err, "database connection error",
			fault.WithContext("host", db.config.Database.Credentials.Host),
		)
	}

	db.logger.Info("Database connected successfully",
		"host", db.config.Database.Credentials.Host,
		"database", db.config.Database.Credentials.Name,
		"pool_max_open", db.config.Database.Pool.MaxOpenConns,
		"pool_max_idle", db.config.Database.Pool.MaxIdleConns,
	)

	return nil
}

// connect performs a single connection attempt
func (db *DB) connect(ctx context.Context) error {
	// Build DSN from configuration
	dsn := db.config.GetDatabaseDSN()

	// Open database connection
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return fault.Wrap(ErrOpenFailed, "sql.Open failed",
			fault.WithWrappedErr(err),
			fault.WithContext("driver", "pgx"),
		)
	}

	// Configure connection pool
	db.configurePool(conn)

	// Verify connection with ping
	pingCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.QueryTimeout)
	defer cancel()

	if err := conn.PingContext(pingCtx); err != nil {
		_ = conn.Close()
		return fault.Wrap(ErrPingFailed, "ping failed",
			fault.WithWrappedErr(err),
			fault.WithContext("timeout", db.config.Database.Connect.QueryTimeout.String()),
		)
	}

	db.conn = conn
	return nil
}

// configurePool configures the connection pool with settings from config
func (db *DB) configurePool(conn *sql.DB) {
	poolConfig := db.config.Database.Pool

	// Maximum number of open connections
	conn.SetMaxOpenConns(poolConfig.MaxOpenConns)

	// Maximum number of idle connections
	conn.SetMaxIdleConns(poolConfig.MaxIdleConns)

	// Maximum lifetime of a connection
	conn.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)

	// Maximum idle time of a connection
	conn.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)
}

// Close gracefully closes the database connection
func (db *DB) Close() error {
	if db.conn == nil {
		return ErrNotConnected
	}

	db.logger.Info("Closing database connection")

	if err := db.conn.Close(); err != nil {
		return fault.Wrap(ErrCloseFailed, "close failed",
			fault.WithWrappedErr(err),
		)
	}

	db.conn = nil
	return nil
}

// Ping verifies database connectivity
func (db *DB) Ping(ctx context.Context) error {
	if db.conn == nil {
		return ErrNotConnected
	}

	pingCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.QueryTimeout)
	defer cancel()

	if err := db.conn.PingContext(pingCtx); err != nil {
		return fault.Wrap(ErrPingFailed, "ping failed",
			fault.WithWrappedErr(err),
			fault.WithContext("timeout", db.config.Database.Connect.QueryTimeout.String()),
		)
	}

	return nil
}

// HealthCheck performs a comprehensive health check
func (db *DB) HealthCheck(ctx context.Context) error {
	if db.conn == nil {
		return ErrNotConnected
	}

	// Check connection with ping
	if err := db.Ping(ctx); err != nil {
		return err
	}

	// Check pool statistics
	stats := db.conn.Stats()

	// Warn if all connections are in use
	if stats.InUse >= stats.MaxOpenConnections {
		db.logger.Warn("All database connections are in use",
			"in_use", stats.InUse,
			"max_open", stats.MaxOpenConnections,
		)
	}

	// Warn if waiting for connections
	if stats.WaitCount > 0 {
		db.logger.Warn("Database connections waiting",
			"wait_count", stats.WaitCount,
			"wait_duration", stats.WaitDuration,
		)
	}

	return nil
}

// Stats returns database connection pool statistics
func (db *DB) Stats() sql.DBStats {
	if db.conn == nil {
		return sql.DBStats{}
	}
	return db.conn.Stats()
}

// DB returns the underlying *sql.DB instance
// Use this for executing queries
func (db *DB) DB() *sql.DB {
	return db.conn
}

// IsConnected returns true if database is connected
func (db *DB) IsConnected() bool {
	return db.conn != nil
}

// ExecContext executes a query without returning rows with timeout
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if db.conn == nil {
		return nil, ErrNotConnected
	}

	execCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.ExecTimeout)
	defer cancel()

	result, err := db.conn.ExecContext(execCtx, query, args...)
	if err != nil {
		db.logger.Error("Query execution failed",
			"query", query,
			"timeout", db.config.Database.Connect.ExecTimeout.String(),
			"error", err.Error(),
		)
		return nil, fault.Wrap(ErrExecFailed, "exec failed",
			fault.WithWrappedErr(err),
			fault.WithContext("query", query),
			fault.WithContext("timeout", db.config.Database.Connect.ExecTimeout.String()),
		)
	}

	return result, nil
}

// QueryContext executes a query that returns rows with timeout
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if db.conn == nil {
		return nil, ErrNotConnected
	}

	queryCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.QueryTimeout)
	defer cancel()

	rows, err := db.conn.QueryContext(queryCtx, query, args...)
	if err != nil {
		db.logger.Error("Query failed",
			"query", query,
			"timeout", db.config.Database.Connect.QueryTimeout.String(),
			"error", err.Error(),
		)
		return nil, fault.Wrap(ErrQueryFailed, "query failed",
			fault.WithWrappedErr(err),
			fault.WithContext("query", query),
			fault.WithContext("timeout", db.config.Database.Connect.QueryTimeout.String()),
		)
	}

	return rows, nil
}

// QueryRowContext executes a query that returns at most one row with timeout
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if db.conn == nil {
		return nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.QueryTimeout)
	defer cancel()

	return db.conn.QueryRowContext(queryCtx, query, args...)
}

// BeginTx starts a transaction with the given options
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if db.conn == nil {
		return nil, ErrNotConnected
	}

	tx, err := db.conn.BeginTx(ctx, opts)
	if err != nil {
		db.logger.Error("Failed to begin transaction", "error", err.Error())
		return nil, fault.Wrap(ErrTransactionFailed, "begin transaction failed",
			fault.WithWrappedErr(err),
		)
	}

	return tx, nil
}

// StartHealthCheckRoutine starts a background goroutine that performs periodic health checks
func (db *DB) StartHealthCheckRoutine(ctx context.Context) {
	if db.conn == nil {
		db.logger.Error("Cannot start health check routine: database not connected")
		return
	}

	period := db.config.Database.Pool.HealthCheckPeriod
	ticker := time.NewTicker(period)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				db.logger.Info("Health check routine stopped")
				return
			case <-ticker.C:
				if err := db.HealthCheck(context.Background()); err != nil {
					db.logger.Error("Health check failed", "error", err)
				} else {
					stats := db.Stats()
					db.logger.Debug("Database health check passed",
						"open_connections", stats.OpenConnections,
						"in_use", stats.InUse,
						"idle", stats.Idle,
					)
				}
			}
		}
	}()

	db.logger.Info("Health check routine started", "period", period)
}
