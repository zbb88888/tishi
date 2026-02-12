package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNormalisePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/healthz", "/healthz"},
		{"/metrics", "/metrics"},
		{"/api/v1/rankings", "/api/v1/rankings"},
		{"/api/v1/projects", "/api/v1/projects"},
		{"/api/v1/projects/123", "/api/v1/projects/:id"},
		{"/api/v1/projects/456/trends", "/api/v1/projects/:id/trends"},
		{"/api/v1/posts", "/api/v1/posts"},
		{"/api/v1/posts/my-weekly-report", "/api/v1/posts/:slug"},
		{"/api/v1/categories", "/api/v1/categories"},
		{"/unknown/path", "/unknown/path"},
	}

	for _, tt := range tests {
		got := normalisePath(tt.input)
		if got != tt.want {
			t.Errorf("normalisePath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMetricsMiddleware(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	// Wrap a simple handler
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rankings", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	// Verify metrics were collected
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gathering metrics: %v", err)
	}

	metricNames := map[string]bool{}
	for _, f := range families {
		metricNames[f.GetName()] = true
	}

	for _, name := range []string{
		"tishi_http_requests_total",
		"tishi_http_request_duration_seconds",
	} {
		if !metricNames[name] {
			t.Errorf("expected metric %q to be registered", name)
		}
	}
}

func TestMetricsMiddleware_SkipsMetricsEndpoint(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	called := false
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected handler to be called for /metrics path")
	}
}

func TestStatusWriter(t *testing.T) {
	rec := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rec, status: http.StatusOK}

	// First WriteHeader sets the status
	sw.WriteHeader(http.StatusNotFound)
	if sw.status != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", sw.status)
	}

	// Second WriteHeader should not change it
	sw.WriteHeader(http.StatusInternalServerError)
	if sw.status != http.StatusNotFound {
		t.Errorf("expected status to remain 404, got %d", sw.status)
	}
}
