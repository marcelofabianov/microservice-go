package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/marcelofabianov/fault"

	"github.com/marcelofabianov/course/pkg/web"
)

func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			done := make(chan struct{})
			var panicVal interface{}

			go func() {
				defer func() {
					if p := recover(); p != nil {
						panicVal = p
					}
					close(done)
				}()

				next.ServeHTTP(w, r.WithContext(ctx))
			}()

			select {
			case <-done:
				if panicVal != nil {
					panic(panicVal)
				}
				return
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					web.Error(w, r, fault.New("request timeout", fault.WithCode(fault.Internal)))
				}
				return
			}
		})
	}
}
