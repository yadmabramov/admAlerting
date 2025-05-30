package gzipmiddleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		if acceptsGzip {
			contentType := r.Header.Get("Content-Type")
			if strings.Contains(contentType, "application/json") ||
				strings.Contains(contentType, "text/html") ||
				strings.Contains(contentType, "text/plain") {

				gz := gzipPool.Get().(*gzip.Writer)
				defer gzipPool.Put(gz)
				defer gz.Close()

				gz.Reset(w)

				w.Header().Set("Content-Encoding", "gzip")
				if contentType != "" {
					w.Header().Set("Content-Type", contentType)
				}

				next.ServeHTTP(&gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.Writer == nil {
		return g.ResponseWriter.Write(b)
	}
	return g.Writer.Write(b)
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	// Ensure Content-Encoding is set before writing headers
	if g.Writer != nil {
		g.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	g.ResponseWriter.WriteHeader(statusCode)
}
