package middleware

import (
	"Alice088/pdf-summarize/prometheus/metrics"
	"net/http"
)

func HttpRequestsInFlight(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.HttpRequestsInFlight.Inc()
		defer metrics.HttpRequestsInFlight.Dec()

		next.ServeHTTP(w, r)
	})
}
