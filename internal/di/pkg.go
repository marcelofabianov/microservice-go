package di

import (
	"context"
	"time"

	"go.uber.org/fx"

	"github.com/marcelofabianov/course/config"
	"github.com/marcelofabianov/course/pkg/cache"
	"github.com/marcelofabianov/course/pkg/database"
	"github.com/marcelofabianov/course/pkg/logger"
	"github.com/marcelofabianov/course/pkg/validation"
)

func ProvideConfig() (*config.Config, error) {
	return config.Load()
}

func ProvideLogger(cfg *config.Config) *logger.Logger {
	return logger.New(&logger.Config{
		Level:       logger.LogLevel(cfg.Logger.Level),
		Format:      logger.FormatJSON,
		ServiceName: cfg.General.ServiceName,
		Environment: cfg.General.Env,
	})
}

func ProvideDatabase(cfg *config.Config, log *logger.Logger, lc fx.Lifecycle) (*database.DB, error) {
	db, err := database.New(cfg)
	if err != nil {
		return nil, err
	}
	db.SetLogger(log.Slog())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			log.Info("connecting to database")
			if err := db.Connect(connectCtx); err != nil {
				return err
			}
			log.Info("database connected successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("closing database connection")
			if err := db.Close(); err != nil {
				log.Error("failed to close database connection", "error", err)
				return err
			}
			log.Info("database connection closed successfully")
			return nil
		},
	})

	return db, nil
}

func ProvideCache(cfg *config.Config, log *logger.Logger, lc fx.Lifecycle) (*cache.Cache, error) {
	cacheClient, err := cache.New(cfg)
	if err != nil {
		return nil, err
	}
	cacheClient.SetLogger(log.Slog())

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			connectCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			log.Info("connecting to cache")
			if err := cacheClient.Connect(connectCtx); err != nil {
				return err
			}
			log.Info("cache connected successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("closing cache connection")
			if err := cacheClient.Close(); err != nil {
				log.Error("failed to close cache connection", "error", err)
				return err
			}
			log.Info("cache connection closed successfully")
			return nil
		},
	})

	return cacheClient, nil
}

func ProvideValidation(log *logger.Logger) validation.Validator {
	return validation.New(log, nil)
}

var PkgModule = fx.Module("pkg",
	fx.Provide(
		ProvideConfig,
		ProvideLogger,
		ProvideDatabase,
		ProvideCache,
		ProvideValidation,
	),
)
