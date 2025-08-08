package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP metrics for the Fortress API
var (
	// HTTPRequestsTotal counts total HTTP requests by method, endpoint, and status
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration measures HTTP request duration distribution
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration distribution",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	// HTTPRequestSize measures HTTP request size distribution
	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "request_size_bytes",
			Help:      "HTTP request size distribution",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8), // 100B to ~100MB
		},
		[]string{"method", "endpoint"},
	)

	// HTTPResponseSize measures HTTP response size distribution
	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "response_size_bytes",
			Help:      "HTTP response size distribution",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8), // 100B to ~100MB
		},
		[]string{"method", "endpoint"},
	)

	// HTTPRequestsInFlight tracks the number of HTTP requests currently being processed
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		},
	)
)

// InitHTTPMetrics initializes HTTP metrics with a custom registry
// This is primarily used for testing to avoid global state conflicts
func InitHTTPMetrics(registry prometheus.Registerer) error {
	HTTPRequestsTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests processed",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration distribution",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	HTTPRequestSize = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "request_size_bytes",
			Help:      "HTTP request size distribution",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	HTTPResponseSize = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "response_size_bytes",
			Help:      "HTTP response size distribution",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	HTTPRequestsInFlight = promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Namespace: "fortress",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		},
	)

	return nil
}