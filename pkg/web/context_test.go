package web

import (
	"context"
	"testing"
)

func TestSetAndGetLogger(t *testing.T) {
	ctx := context.Background()

	defaultLogger := GetLogger(ctx)
	if defaultLogger == nil {
		t.Error("expected default logger, got nil")
	}
}

func TestSetAndGetRequestID(t *testing.T) {
	ctx := context.Background()

	requestID := GetRequestID(ctx)
	if requestID != "" {
		t.Errorf("expected empty request ID, got %s", requestID)
	}

	ctx = SetRequestID(ctx, "test-request-id")
	requestID = GetRequestID(ctx)
	if requestID != "test-request-id" {
		t.Errorf("expected test-request-id, got %s", requestID)
	}
}
