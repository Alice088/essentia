package prometheus

import (
	"Alice088/pdf-summarize/internal/prometheus/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func UpMetrics() {
	prometheus.MustRegister(metrics.HttpRequestsTotal)
	prometheus.MustRegister(metrics.HttpRequestsInFlight)
	prometheus.MustRegister(metrics.HttpRequestTotalDuration)
}
