package logmiddleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)

			logger.Info("request processed",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", wrapped.status),
				zap.Int("size", wrapped.size),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	status int
	size   int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{ResponseWriter: w}
}

func (w *responseWriterWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}
