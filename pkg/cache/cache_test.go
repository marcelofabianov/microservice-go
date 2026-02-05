package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/marcelofabianov/course/config"
	"github.com/marcelofabianov/course/pkg/cache"
	"github.com/marcelofabianov/course/pkg/logger"
)

func TestNew(t *testing.T) {
	t.Run("creates cache instance with valid config", func(t *testing.T) {
		cfg, err := config.Load()
		if err != nil {
			t.Skip("Config not available")
		}

		c, err := cache.New(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if c == nil {
			t.Fatal("expected cache instance, got nil")
		}
	})

	t.Run("returns error with nil config", func(t *testing.T) {
		c, err := cache.New(nil)
		if err == nil {
			t.Fatal("expected error for nil config, got nil")
		}

		if c != nil {
			t.Fatal("expected nil cache, got instance")
		}
	})
}

func TestCache_SetLogger(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Skip("Config not available")
	}

	c, _ := cache.New(cfg)

	log := logger.New(&logger.Config{
		Level:       logger.LevelDebug,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	c.SetLogger(log.Slog())

	if c.IsConnected() {
		t.Error("cache should not be connected")
	}
}

func TestCache_BasicOperations(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Skip("Config not available")
	}

	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	c, _ := cache.New(cfg)
	c.SetLogger(log.Slog())

	t.Run("operations fail when not connected", func(t *testing.T) {
		ctx := context.Background()

		err := c.Set(ctx, "test", "value", time.Minute)
		if err == nil {
			t.Error("expected error when not connected")
		}

		_, err = c.Get(ctx, "test")
		if err == nil {
			t.Error("expected error when not connected")
		}

		err = c.Delete(ctx, "test")
		if err == nil {
			t.Error("expected error when not connected")
		}
	})

	t.Run("IsConnected returns false initially", func(t *testing.T) {
		if c.IsConnected() {
			t.Error("cache should not be connected initially")
		}
	})
}

func TestCache_Errors(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Skip("Config not available")
	}

	c, _ := cache.New(cfg)
	ctx := context.Background()

	t.Run("Ping fails when not connected", func(t *testing.T) {
		err := c.Ping(ctx)
		if err == nil {
			t.Error("expected error when not connected")
		}
	})

	t.Run("HealthCheck fails when not connected", func(t *testing.T) {
		err := c.HealthCheck(ctx)
		if err == nil {
			t.Error("expected error when not connected")
		}
	})

	t.Run("Close fails when not connected", func(t *testing.T) {
		err := c.Close()
		if err == nil {
			t.Error("expected error when not connected")
		}
	})
}

func TestCache_Stats(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Skip("Config not available")
	}

	c, _ := cache.New(cfg)

	t.Run("returns empty stats when not connected", func(t *testing.T) {
		stats := c.Stats()
		if stats == nil {
			t.Error("expected stats, got nil")
		}
	})
}

func TestCache_Client(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Skip("Config not available")
	}

	c, _ := cache.New(cfg)

	t.Run("returns nil client when not connected", func(t *testing.T) {
		client := c.Client()
		if client != nil {
			t.Error("expected nil client when not connected")
		}
	})
}
