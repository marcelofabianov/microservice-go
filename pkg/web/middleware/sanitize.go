package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/microcosm-cc/bluemonday"
)

func SanitizeMiddleware(next http.Handler) http.Handler {
	policy := bluemonday.StrictPolicy()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldSanitize(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		if r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		sanitized := policy.SanitizeBytes(body)

		r.Body = io.NopCloser(bytes.NewReader(sanitized))
		r.ContentLength = int64(len(sanitized))

		next.ServeHTTP(w, r)
	})
}

func shouldSanitize(method string) bool {
	return method == http.MethodPost ||
		method == http.MethodPut ||
		method == http.MethodPatch
}
