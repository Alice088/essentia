package middleware

import (
	"Alice088/essentia/pkg/prometheus/metrics"
	"net/http"
	"time"
)

func PrometheusHttpRequestTotalDuration(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()

		metrics.HttpRequestTotalDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)
	})
}
