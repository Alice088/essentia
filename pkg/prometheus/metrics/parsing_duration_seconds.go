package metrics

import "github.com/prometheus/client_golang/prometheus"

var ParsingDurationSeconds = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "parsing_duration_seconds",
		Help:    "PDF parsing processing duration per file",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 20, 30, 60, 120},
	},
	[]string{"status"},
)
