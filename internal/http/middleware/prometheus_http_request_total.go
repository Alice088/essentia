package middleware

import (
	"Alice088/pdf-summarize/prometheus/metrics"
	"net/http"
)

func PrometheusHttpRequestTotal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.HttpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Inc()

		next.ServeHTTP(w, r)
	})
}
