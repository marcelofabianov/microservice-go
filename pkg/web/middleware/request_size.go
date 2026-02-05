package middleware

import (
	"net/http"

	"github.com/marcelofabianov/fault"

	"github.com/marcelofabianov/course/pkg/web"
)

func RequestSize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				err := fault.New(
					"request body too large",
					fault.WithCode(fault.Invalid),
					fault.WithContext("max_bytes", maxBytes),
					fault.WithContext("content_length", r.ContentLength),
				)
				web.Error(w, r, err)
				return
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
