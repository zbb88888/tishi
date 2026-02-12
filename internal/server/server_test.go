package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/zbb88888/tishi/internal/config"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	log, _ := zap.NewDevelopment()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
	}
	// pool 为 nil — 只测试不需要 DB 的 handler
	return New(nil, log, cfg)
}

func TestHealthzEndpoint(t *testing.T) {
	srv := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}

	// DB ping will fail (pool is nil) — database should be "disconnected"
	if resp["database"] != "disconnected" {
		t.Errorf("expected database disconnected, got %v", resp["database"])
	}
}

func TestParsePagination_Defaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	page, perPage, offset := parsePagination(req)

	if page != 1 {
		t.Errorf("expected page 1, got %d", page)
	}
	if perPage != 20 {
		t.Errorf("expected perPage 20, got %d", perPage)
	}
	if offset != 0 {
		t.Errorf("expected offset 0, got %d", offset)
	}
}

func TestParsePagination_Custom(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=3&per_page=50", nil)
	page, perPage, offset := parsePagination(req)

	if page != 3 {
		t.Errorf("expected page 3, got %d", page)
	}
	if perPage != 50 {
		t.Errorf("expected perPage 50, got %d", perPage)
	}
	if offset != 100 {
		t.Errorf("expected offset 100, got %d", offset)
	}
}

func TestParsePagination_PerPageCap(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?per_page=200", nil)
	_, perPage, _ := parsePagination(req)

	// per_page > 100 should be ignored, keeping default 20
	if perPage != 20 {
		t.Errorf("expected perPage 20 (cap), got %d", perPage)
	}
}

func TestParsePagination_Invalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=-1&per_page=abc", nil)
	page, perPage, _ := parsePagination(req)

	if page != 1 {
		t.Errorf("expected page 1 (default), got %d", page)
	}
	if perPage != 20 {
		t.Errorf("expected perPage 20 (default), got %d", perPage)
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusCreated, map[string]string{"foo": "bar"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Errorf("unexpected content type: %s", ct)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var resp apiError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding: %v", err)
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got %s", resp.Error.Code)
	}
}
