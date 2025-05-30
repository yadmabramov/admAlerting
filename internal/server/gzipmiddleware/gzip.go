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

		if r.URL.Path == "/ping" {
			next.ServeHTTP(w, r)
			return
		}

		if acceptsGzip {
			gz := gzipPool.Get().(*gzip.Writer)
			defer gzipPool.Put(gz)
			defer gz.Close()

			gz.Reset(w)

			w.Header().Set("Content-Encoding", "gzip")

			next.ServeHTTP(&gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
			return
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
	if g.Writer != nil {
		g.ResponseWriter.Header().Del("Content-Length")
		g.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	g.ResponseWriter.WriteHeader(statusCode)
}
