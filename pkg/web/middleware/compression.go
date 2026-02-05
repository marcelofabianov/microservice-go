package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/marcelofabianov/course/pkg/web"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer        io.Writer
	wroteHeader   bool
	statusCode    int
	headerWritten bool
	contentType   string
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.statusCode = statusCode

	// Set Content-Encoding only when actually writing compressed content
	if !w.headerWritten {
		// Preserve original Content-Type if set
		if ct := w.ResponseWriter.Header().Get("Content-Type"); ct != "" {
			w.contentType = ct
		}

		w.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w.ResponseWriter.Header().Set("Vary", "Accept-Encoding")
		w.ResponseWriter.Header().Del("Content-Length")

		// Restore Content-Type if it was set
		if w.contentType != "" {
			w.ResponseWriter.Header().Set("Content-Type", w.contentType)
		}

		w.headerWritten = true
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		// Capture Content-Type before WriteHeader
		if ct := w.ResponseWriter.Header().Get("Content-Type"); ct != "" {
			w.contentType = ct
		}
		w.WriteHeader(http.StatusOK)
	}
	return w.Writer.Write(b)
}

func Compression(level int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Get gzip writer from pool
			gz := gzipWriterPool.Get().(*gzip.Writer)
			defer gzipWriterPool.Put(gz)

			gz.Reset(w)
			defer func() {
				if err := gz.Close(); err != nil {
					log := web.GetLogger(r.Context())
					log.Error("failed to close gzip writer", "error", err)
				}
			}()

			// Set compression level if valid
			if level >= gzip.DefaultCompression && level <= gzip.BestCompression {
				_ = gz.Close()
				gz, _ = gzip.NewWriterLevel(w, level)
			}

			// Wrap response writer
			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				Writer:         gz,
			}

			next.ServeHTTP(gzw, r)
		})
	}
}
