//go:build integration
// +build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/marcelofabianov/course/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_Connect_Integration(t *testing.T) {
	cfg, err := config.Load()
	require.NoError(t, err)

	db, err := New(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("connects successfully to database", func(t *testing.T) {
		err := db.Connect(ctx)
		defer db.Close()

		assert.NoError(t, err)
		assert.True(t, db.IsConnected())
		assert.NotNil(t, db.DB())
	})

	t.Run("ping succeeds after connection", func(t *testing.T) {
		db2, _ := New(cfg)
		err := db2.Connect(ctx)
		require.NoError(t, err)
		defer db2.Close()

		err = db2.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("health check succeeds after connection", func(t *testing.T) {
		db3, _ := New(cfg)
		err := db3.Connect(ctx)
		require.NoError(t, err)
		defer db3.Close()

		err = db3.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("executes query successfully", func(t *testing.T) {
		db4, _ := New(cfg)
		err := db4.Connect(ctx)
		require.NoError(t, err)
		defer db4.Close()

		rows, err := db4.QueryContext(ctx, "SELECT 1 as num, 'test' as txt")
		require.NoError(t, err)
		defer rows.Close()

		assert.True(t, rows.Next())

		var num int
		var txt string
		err = rows.Scan(&num, &txt)

		assert.NoError(t, err)
		assert.Equal(t, 1, num)
		assert.Equal(t, "test", txt)
	})

	t.Run("executes statement successfully", func(t *testing.T) {
		db5, _ := New(cfg)
		err := db5.Connect(ctx)
		require.NoError(t, err)
		defer db5.Close()

		// Create temporary table
		_, err = db5.ExecContext(ctx, `
			CREATE TEMPORARY TABLE test_table (
				id SERIAL PRIMARY KEY,
				name TEXT NOT NULL
			)
		`)
		require.NoError(t, err)

		// Insert data
		result, err := db5.ExecContext(ctx, "INSERT INTO test_table (name) VALUES ($1)", "test")
		require.NoError(t, err)

		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)
	})

	t.Run("transaction works correctly", func(t *testing.T) {
		db6, _ := New(cfg)
		err := db6.Connect(ctx)
		require.NoError(t, err)
		defer db6.Close()

		// Begin transaction
		tx, err := db6.BeginTx(ctx, nil)
		require.NoError(t, err)

		// Rollback transaction
		err = tx.Rollback()
		assert.NoError(t, err)
	})

	t.Run("returns pool statistics", func(t *testing.T) {
		db7, _ := New(cfg)
		err := db7.Connect(ctx)
		require.NoError(t, err)
		defer db7.Close()

		stats := db7.Stats()

		assert.GreaterOrEqual(t, stats.MaxOpenConnections, 1)
		assert.GreaterOrEqual(t, stats.OpenConnections, 0)
	})
}
