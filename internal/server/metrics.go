// Package server provides HTTP API server metrics via Prometheus.
package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds Prometheus collectors for HTTP request observation.
type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestsActive  prometheus.Gauge
}

// NewMetrics registers and returns a new Metrics instance.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	factory := promauto.With(reg)

	return &Metrics{
		requestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "tishi",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests processed.",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "tishi",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "Histogram of HTTP request durations.",
				Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			},
			[]string{"method", "path"},
		),
		requestsActive: factory.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "tishi",
				Subsystem: "http",
				Name:      "requests_active",
				Help:      "Number of in-flight HTTP requests.",
			},
		),
	}
}

// Middleware returns an http.Handler middleware that records metrics.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip metrics endpoint itself to avoid recursion
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		m.requestsActive.Inc()
		start := time.Now()

		// Wrap response writer to capture status code
		ww := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)

		// Normalise path for cardinality control
		path := normalisePath(r.URL.Path)
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(ww.status)

		m.requestDuration.WithLabelValues(r.Method, path).Observe(duration)
		m.requestsTotal.WithLabelValues(r.Method, path, status).Inc()
		m.requestsActive.Dec()
	})
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (sw *statusWriter) WriteHeader(code int) {
	if !sw.wroteHeader {
		sw.status = code
		sw.wroteHeader = true
	}
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if !sw.wroteHeader {
		sw.wroteHeader = true
	}
	return sw.ResponseWriter.Write(b)
}

// normalisePath reduces URL path cardinality for Prometheus labels.
// /api/v1/projects/123 → /api/v1/projects/:id
// /api/v1/posts/some-slug → /api/v1/posts/:slug
func normalisePath(p string) string {
	switch {
	case p == "/healthz" || p == "/metrics":
		return p
	case p == "/api/v1/rankings" || p == "/api/v1/projects" ||
		p == "/api/v1/posts" || p == "/api/v1/categories":
		return p
	case len(p) > len("/api/v1/projects/") && p[:len("/api/v1/projects/")] == "/api/v1/projects/":
		// /api/v1/projects/{id} or /api/v1/projects/{id}/trends
		suffix := p[len("/api/v1/projects/"):]
		for i, c := range suffix {
			if c == '/' {
				return "/api/v1/projects/:id" + suffix[i:]
			}
		}
		return "/api/v1/projects/:id"
	case len(p) > len("/api/v1/posts/") && p[:len("/api/v1/posts/")] == "/api/v1/posts/":
		return "/api/v1/posts/:slug"
	default:
		return p
	}
}
