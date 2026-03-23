package metrics

import "github.com/prometheus/client_golang/prometheus"

var HttpRequestsInFlight = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current number of HTTP requests being processed",
	},
)
