package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marcelofabianov/course/pkg/web/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeMiddleware_RemovesScriptTags(t *testing.T) {
	handler := middleware.SanitizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	}))

	maliciousJSON := `{"name":"<script>alert('xss')</script>John"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(maliciousJSON))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.NotContains(t, body, "<script>")
	assert.NotContains(t, body, "alert")
}

func TestSanitizeMiddleware_RemovesHTMLTags(t *testing.T) {
	handler := middleware.SanitizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	}))

	maliciousJSON := `{"bio":"<b>Bold</b> text with <img src=x onerror=alert(1)>"}`
	req := httptest.NewRequest(http.MethodPost, "/api/profile", bytes.NewBufferString(maliciousJSON))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.NotContains(t, body, "<b>")
	assert.NotContains(t, body, "<img")
	assert.NotContains(t, body, "onerror")
}

func TestSanitizeMiddleware_AllowsCleanContent(t *testing.T) {
	handler := middleware.SanitizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	}))

	cleanJSON := `{"name":"John Doe","email":"john@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(cleanJSON))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.Contains(t, body, "John Doe")
	assert.Contains(t, body, "john@example.com")
}

func TestSanitizeMiddleware_OnlyProcessesMutatingMethods(t *testing.T) {
	handler := middleware.SanitizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Write(body)
	}))

	tests := []struct {
		method         string
		shouldSanitize bool
	}{
		{http.MethodGet, false},
		{http.MethodPost, true},
		{http.MethodPut, true},
		{http.MethodPatch, true},
		{http.MethodDelete, false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			malicious := `{"data":"<script>alert(1)</script>"}`
			req := httptest.NewRequest(tt.method, "/api/test", bytes.NewBufferString(malicious))

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if tt.shouldSanitize {
				assert.NotContains(t, rec.Body.String(), "<script>")
			}
		})
	}
}

func TestSanitizeMiddleware_HandlesEmptyBody(t *testing.T) {
	handler := middleware.SanitizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	rec := httptest.NewRecorder()

	require.NotPanics(t, func() {
		handler.ServeHTTP(rec, req)
	})

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSanitizeMiddleware_PreventsSQLInjection(t *testing.T) {
	handler := middleware.SanitizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Write(body)
	}))

	maliciousJSON := `{"query":"<script>'; DROP TABLE users; --</script>"}`
	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBufferString(maliciousJSON))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.NotContains(t, body, "<script>")
}
