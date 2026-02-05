package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/marcelofabianov/course/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("creates database instance with valid config", func(t *testing.T) {
		cfg, err := config.Load()
		require.NoError(t, err)

		db, err := New(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, db.config)
		assert.NotNil(t, db.logger)
		assert.Nil(t, db.conn)
		assert.False(t, db.IsConnected())
	})

	t.Run("returns error with nil config", func(t *testing.T) {
		db, err := New(nil)

		assert.True(t, errors.Is(err, ErrInvalidConfig))
		assert.Nil(t, db)
	})
}

func TestDB_Connect(t *testing.T) {
	t.Run("returns error when already connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		// Simulate already connected
		db.conn = &sql.DB{}

		err := db.Connect(context.Background())

		assert.True(t, errors.Is(err, ErrAlreadyConnected))
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := db.Connect(ctx)

		assert.Error(t, err)
	})

	t.Run("respects context timeout", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		// Very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		err := db.Connect(ctx)

		assert.Error(t, err)
	})
}

func TestDB_Close(t *testing.T) {
	t.Run("returns error when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		err := db.Close()

		assert.True(t, errors.Is(err, ErrNotConnected))
	})
}

func TestDB_Ping(t *testing.T) {
	t.Run("returns error when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		err := db.Ping(context.Background())

		assert.True(t, errors.Is(err, ErrNotConnected))
	})
}

func TestDB_HealthCheck(t *testing.T) {
	t.Run("returns error when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		err := db.HealthCheck(context.Background())

		assert.True(t, errors.Is(err, ErrNotConnected))
	})
}

func TestDB_Stats(t *testing.T) {
	t.Run("returns empty stats when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		stats := db.Stats()

		assert.Equal(t, 0, stats.OpenConnections)
		assert.Equal(t, 0, stats.InUse)
		assert.Equal(t, 0, stats.Idle)
	})
}

func TestDB_DB(t *testing.T) {
	t.Run("returns nil when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		conn := db.DB()

		assert.Nil(t, conn)
	})
}

func TestDB_ExecContext(t *testing.T) {
	t.Run("returns error when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		_, err := db.ExecContext(context.Background(), "SELECT 1")

		assert.True(t, errors.Is(err, ErrNotConnected))
	})
}

func TestDB_QueryContext(t *testing.T) {
	t.Run("returns error when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		_, err := db.QueryContext(context.Background(), "SELECT 1")

		assert.True(t, errors.Is(err, ErrNotConnected))
	})
}

func TestDB_QueryRowContext(t *testing.T) {
	t.Run("returns nil when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		row := db.QueryRowContext(context.Background(), "SELECT 1")

		assert.Nil(t, row)
	})
}

func TestDB_BeginTx(t *testing.T) {
	t.Run("returns error when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		_, err := db.BeginTx(context.Background(), nil)

		assert.True(t, errors.Is(err, ErrNotConnected))
	})
}

func TestDB_IsConnected(t *testing.T) {
	t.Run("returns false when not connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		assert.False(t, db.IsConnected())
	})

	t.Run("returns true when connected", func(t *testing.T) {
		cfg, _ := config.Load()
		db, _ := New(cfg)

		// Simulate connection
		db.conn = &sql.DB{}

		assert.True(t, db.IsConnected())
	})
}
