package middleware

import (
	"Alice088/essentia/pkg/prometheus/metrics"
	"net/http"
	"strconv"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func PrometheusHttpRequestTotal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusHook := &statusWriter{ResponseWriter: w, status: 200}

		next.ServeHTTP(statusHook, r)

		metrics.HttpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			strconv.Itoa(statusHook.status),
		).Inc()
	})
}
