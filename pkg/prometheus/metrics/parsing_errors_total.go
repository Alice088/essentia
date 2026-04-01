package metrics

import "github.com/prometheus/client_golang/prometheus"

var ParsingErrorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "parsing_errors_total",
		Help: "total number of parsing errors by type",
	},
	[]string{"error_type"},
)
