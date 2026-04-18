package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}

	if m.GetUptime() == 0 {
		t.Fatal("Uptime should be greater than 0")
	}
}

func TestIncrementRequest(t *testing.T) {
	m := NewMetrics()

	m.IncrementRequest("/api/login")

	count := m.GetRequestCount("/api/login")
	if count != 1 {
		t.Errorf("Expected request count 1, got %d", count)
	}

	m.IncrementRequest("/api/login")
	m.IncrementRequest("/api/login")

	count = m.GetRequestCount("/api/login")
	if count != 3 {
		t.Errorf("Expected request count 3, got %d", count)
	}
}

func TestIncrementError(t *testing.T) {
	m := NewMetrics()

	m.IncrementError("/api/login")

	count := m.GetErrorCount("/api/login")
	if count != 1 {
		t.Errorf("Expected error count 1, got %d", count)
	}
}

func TestSetDeviceMetrics(t *testing.T) {
	m := NewMetrics()

	m.SetActiveDevices(5)
	m.SetTotalDevices(10)

	active := m.GetActiveDevices()
	total := m.GetTotalDevices()

	if active != 5 {
		t.Errorf("Expected active devices 5, got %d", active)
	}

	if total != 10 {
		t.Errorf("Expected total devices 10, got %d", total)
	}
}

func TestGetAllMetrics(t *testing.T) {
	m := NewMetrics()

	m.IncrementRequest("/api/login")
	m.IncrementRequest("/api/devices")
	m.IncrementError("/api/login")
	m.SetActiveDevices(3)
	m.SetTotalDevices(5)

	allMetrics := m.GetAllMetrics()

	if allMetrics == nil {
		t.Fatal("GetAllMetrics returned nil")
	}

	requests, ok := allMetrics["requests"].(map[string]int64)
	if !ok {
		t.Fatal("requests is not a map[string]int64")
	}

	if requests["/api/login"] != 1 {
		t.Errorf("Expected /api/login requests 1, got %d", requests["/api/login"])
	}

	if allMetrics["active_devices"].(int64) != 3 {
		t.Errorf("Expected active devices 3, got %v", allMetrics["active_devices"])
	}

	if allMetrics["total_devices"].(int64) != 5 {
		t.Errorf("Expected total devices 5, got %v", allMetrics["total_devices"])
	}
}

func TestMetricsHandler(t *testing.T) {
	m := NewMetrics()

	m.IncrementRequest("/api/login")
	m.SetActiveDevices(2)

	handler := m.Handler()

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected content-type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestHealthHandler(t *testing.T) {
	m := NewMetrics()

	handler := m.Handler()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMetricsHandlerNotFound(t *testing.T) {
	m := NewMetrics()

	handler := m.Handler()

	req := httptest.NewRequest("GET", "/unknown", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestGetUptime(t *testing.T) {
	m := NewMetrics()

	time.Sleep(10 * time.Millisecond)

	uptime := m.GetUptime()
	if uptime < 10*time.Millisecond {
		t.Errorf("Expected uptime at least 10ms, got %v", uptime)
	}
}
