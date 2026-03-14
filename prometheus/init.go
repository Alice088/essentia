package prometheus

import (
	"Alice088/pdf-summarize/prometheus/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func UpMetrics() {
	prometheus.MustRegister(metrics.HttpRequestsTotal)
}
