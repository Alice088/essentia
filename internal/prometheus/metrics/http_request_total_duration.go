package metrics

import "github.com/prometheus/client_golang/prometheus"

var HttpRequestTotalDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_total_duration_seconds",
		Help:    "Total HTTP request processing time",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "endpoint"},
)
