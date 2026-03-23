package prometheus

import (
	metrics2 "Alice088/essentia/pkg/prometheus/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

func UpMetrics() {
	prometheus.MustRegister(metrics2.HttpRequestsTotal)
	prometheus.MustRegister(metrics2.HttpRequestsInFlight)
	prometheus.MustRegister(metrics2.HttpRequestTotalDuration)
}
