package metrics

import "github.com/prometheus/client_golang/prometheus"

var ParsingPDFSizeBytes = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "parsing_pdf_size_bytes",
		Help:    "PDF file size in bytes processed by parsing worker",
		Buckets: []float64{1024, 4096, 16384, 65536, 262144, 524288, 1048576, 2097152, 3145728, 4194304, 5242880},
	},
	[]string{"status"},
)
