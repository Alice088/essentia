package metrics

import "github.com/prometheus/client_golang/prometheus"

// ParsingStatus = Failed | Success
type ParsingStatus = string

const (
	Failed  ParsingStatus = "failed"
	Success ParsingStatus = "success"
)

var ParsingTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "parsing_total",
		Help: "total number of PDF parsing tasks",
	},
	[]ParsingStatus{"status"},
)
