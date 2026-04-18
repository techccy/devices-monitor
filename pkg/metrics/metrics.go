package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	requests      map[string]int64
	errors        map[string]int64
	activeDevices int64
	totalDevices  int64

	startTime time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		requests:  make(map[string]int64),
		errors:    make(map[string]int64),
		startTime: time.Now(),
	}
}

func (m *Metrics) IncrementRequest(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[endpoint]++
}

func (m *Metrics) IncrementError(endpoint string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[endpoint]++
}

func (m *Metrics) SetActiveDevices(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeDevices = count
}

func (m *Metrics) SetTotalDevices(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalDevices = count
}

func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

func (m *Metrics) GetRequestCount(endpoint string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requests[endpoint]
}

func (m *Metrics) GetErrorCount(endpoint string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errors[endpoint]
}

func (m *Metrics) GetActiveDevices() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activeDevices
}

func (m *Metrics) GetTotalDevices() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalDevices
}

func (m *Metrics) GetAllMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	requestsCopy := make(map[string]int64)
	errorsCopy := make(map[string]int64)

	for k, v := range m.requests {
		requestsCopy[k] = v
	}
	for k, v := range m.errors {
		errorsCopy[k] = v
	}

	return map[string]interface{}{
		"uptime":         m.GetUptime().String(),
		"requests":       requestsCopy,
		"errors":         errorsCopy,
		"active_devices": m.activeDevices,
		"total_devices":  m.totalDevices,
		"total_requests": m.sumMap(m.requests),
		"total_errors":   m.sumMap(m.errors),
	}
}

func (m *Metrics) sumMap(mmap map[string]int64) int64 {
	total := int64(0)
	for _, v := range mmap {
		total += v
	}
	return total
}

func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			w.Header().Set("Content-Type", "application/json")
			metrics := m.GetAllMetrics()

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{\"metrics\":"))

			if err := json.NewEncoder(w).Encode(metrics); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else if r.URL.Path == "/health" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy"}`))
		} else {
			http.NotFound(w, r)
		}
	})
}
