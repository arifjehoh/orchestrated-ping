package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/arifjehoh/orchestrated-ping/internal/metrics"
	"github.com/go-chi/chi/v5"
)

func Metrics() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start).Seconds()
			endpoint := chi.RouteContext(r.Context()).RoutePattern()
			statusCode := strconv.Itoa(ww.statusCode)

			metrics.HttpDuration.WithLabelValues(r.Method, endpoint, statusCode).Observe(duration)
			metrics.HttpRequestsTotal.WithLabelValues(r.Method, endpoint, statusCode).Inc()
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
