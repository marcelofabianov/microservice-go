package web

import (
	"context"

	"github.com/marcelofabianov/course/pkg/logger"
)

type contextKey string

const (
	LoggerCtxKey    contextKey = "logger"
	RequestIDCtxKey contextKey = "request_id"
)

func GetLogger(ctx context.Context) *logger.Logger {
	if log, ok := ctx.Value(LoggerCtxKey).(*logger.Logger); ok {
		return log
	}
	return logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatJSON,
		ServiceName: "default",
		Environment: "development",
	})
}

func SetLogger(ctx context.Context, log *logger.Logger) context.Context {
	return context.WithValue(ctx, LoggerCtxKey, log)
}

func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDCtxKey).(string); ok {
		return id
	}
	return ""
}

func SetRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDCtxKey, id)
}
