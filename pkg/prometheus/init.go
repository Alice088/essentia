package prometheus

import (
	"Alice088/essentia/pkg/prometheus/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func UpMetrics() {
	prometheus.MustRegister(metrics.HttpRequestsTotal)
	prometheus.MustRegister(metrics.HttpRequestsInFlight)
	prometheus.MustRegister(metrics.HttpRequestTotalDuration)
	prometheus.MustRegister(metrics.ParsingTotal)
	prometheus.MustRegister(metrics.ParsingDurationSeconds)
	prometheus.MustRegister(metrics.ParsingPDFSizeBytes)
}
