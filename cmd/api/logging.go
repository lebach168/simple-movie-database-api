package main

import (
	"fmt"
	"net/http"
	"time"
)

type WrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *WrappedWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (app *application) LoggingHTTPHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := &WrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(writer, r)
		app.logger.Info(fmt.Sprintf("%d %s %s %s", writer.statusCode, r.Method, r.URL.Path, time.Since(start)))
	})
}
