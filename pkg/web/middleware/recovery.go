package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/marcelofabianov/fault"

	"github.com/marcelofabianov/course/pkg/logger"
	"github.com/marcelofabianov/course/pkg/web"
)

func Recovery(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					stack := debug.Stack()

					err := fault.New(
						fmt.Sprintf("panic recovered: %v", rvr),
						fault.WithCode(fault.Internal),
						fault.WithContext("stack", string(stack)),
					)

					log.Error("panic recovered",
						"panic", rvr,
						"path", r.URL.Path,
						"method", r.Method,
						"error", err.Error(),
						"stack", string(stack), // Add stack trace to logs
					)

					web.InternalServerError(w, r, fault.New("internal server error", fault.WithCode(fault.Internal)))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
