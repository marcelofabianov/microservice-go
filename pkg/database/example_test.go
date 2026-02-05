package database_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/marcelofabianov/course/config"
	"github.com/marcelofabianov/course/pkg/database"
)

// Example_basicConnection demonstrates basic database connection
func Example_basicConnection() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Create database instance
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Connect with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Connected to database successfully")
}

// Example_withRetry demonstrates connection with automatic retry
func Example_withRetry() {
	cfg, _ := config.Load()
	db, _ := database.New(cfg)

	// Connection automatically retries on failure using config settings:
	// - BackoffMin: 200ms
	// - BackoffMax: 15s
	// - BackoffRetries: 7
	// - Exponential backoff with jitter

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Printf("Connection failed after retries: %v", err)
		return
	}
	defer db.Close()

	fmt.Println("Connected with automatic retry")
}

// Example_query demonstrates executing a query
func Example_query() {
	cfg, _ := config.Load()
	db, _ := database.New(cfg)

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Query with automatic timeout from config
	rows, err := db.QueryContext(ctx, "SELECT id, name FROM users LIMIT 10")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User: %d - %s\n", id, name)
	}
}

// Example_transaction demonstrates using transactions
func Example_transaction() {
	cfg, _ := config.Load()
	db, _ := database.New(cfg)

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Execute statements in transaction
	_, err = tx.ExecContext(ctx, "INSERT INTO users (name, email) VALUES ($1, $2)",
		"John Doe", "john@example.com")
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Transaction committed successfully")
}

// Example_healthCheck demonstrates health check usage
func Example_healthCheck() {
	cfg, _ := config.Load()
	db, _ := database.New(cfg)

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Perform health check
	if err := db.HealthCheck(ctx); err != nil {
		log.Printf("Health check failed: %v", err)
		return
	}

	// Get pool statistics
	stats := db.Stats()
	fmt.Printf("Pool stats: Open=%d, InUse=%d, Idle=%d\n",
		stats.OpenConnections, stats.InUse, stats.Idle)
}

// Example_backgroundHealthCheck demonstrates periodic health checks
func Example_backgroundHealthCheck() {
	cfg, _ := config.Load()
	db, _ := database.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Start background health check routine
	// Checks every HealthCheckPeriod (from config, default: 1m)
	db.StartHealthCheckRoutine(ctx)

	// Application continues running...
	// Health checks run in background

	fmt.Println("Health check routine started")
}

// Example_gracefulShutdown demonstrates graceful shutdown
func Example_gracefulShutdown() {
	cfg, _ := config.Load()
	db, _ := database.New(cfg)

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}

	// Simulate application shutdown
	fmt.Println("Shutting down...")

	// Close database connection gracefully
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	fmt.Println("Database closed successfully")
}

// Example_poolConfiguration demonstrates connection pool settings
func Example_poolConfiguration() {
	cfg, _ := config.Load()
	db, _ := database.New(cfg)

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Pool is automatically configured from config:
	// - MaxOpenConns: 20 (max concurrent connections)
	// - MaxIdleConns: 10 (max idle connections in pool)
	// - ConnMaxLifetime: 15m (connection lifetime)
	// - ConnMaxIdleTime: 5m (idle connection lifetime)

	stats := db.Stats()
	fmt.Printf("Max Open Connections: %d\n", stats.MaxOpenConnections)
	fmt.Printf("Currently Open: %d\n", stats.OpenConnections)
	fmt.Printf("In Use: %d\n", stats.InUse)
	fmt.Printf("Idle: %d\n", stats.Idle)
}

// Example_withCustomLogger demonstrates using custom logger
func Example_withCustomLogger() {
	cfg, _ := config.Load()
	db, _ := database.New(cfg)

	// Create custom logger from our logger package
	// logger := logger.NewFromAppConfig(cfg.Logger.Level, cfg.General.ServiceName, cfg.General.Env)
	// db.SetLogger(logger.Slog())

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Using custom logger")
}

// Example_rawSQLDB demonstrates accessing raw sql.DB
func Example_rawSQLDB() {
	cfg, _ := config.Load()
	db, _ := database.New(cfg)

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Get raw *sql.DB for use with other libraries (sqlx, squirrel, etc.)
	rawDB := db.DB()

	// Use with sqlx, gorm, or any library that accepts *sql.DB
	_ = rawDB

	fmt.Println("Raw sql.DB access available")
}
