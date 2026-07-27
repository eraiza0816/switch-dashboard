package server

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "switchd_http_requests_total",
			Help: "Total number of HTTP requests by method, path, and status.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "switchd_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	activeSwitches = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "switchd_active_switches",
			Help: "Number of active (enabled) switches in config.",
		},
	)

	duckdbReady = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "switchd_duckdb_ready",
			Help: "DuckDB database readiness (1 = ready, 0 = not ready).",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(activeSwitches)
	prometheus.MustRegister(duckdbReady)
}

func metricsHandler() http.Handler {
	return promhttp.Handler()
}

func recordMetrics(method, path string, status int, durationSeconds float64) {
	httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(durationSeconds)
}

func SetActiveSwitchesGauge(n int) {
	activeSwitches.Set(float64(n))
}

func SetDuckDBReadyGauge(ready bool) {
	if ready {
		duckdbReady.Set(1)
	} else {
		duckdbReady.Set(0)
	}
}
