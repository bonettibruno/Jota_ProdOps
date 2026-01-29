package core

import "sync"

// Metrics stores operational telemetry for the platform
type Metrics struct {
	mu              sync.Mutex
	TotalRequests   int            `json:"total_requests"`
	TotalHandoffs   int            `json:"total_handoffs"`
	TotalEscalates  int            `json:"total_escalates"`
	RequestsByAgent map[string]int `json:"requests_by_agent"`
}

var globalMetrics = &Metrics{
	RequestsByAgent: make(map[string]int),
}

// GetMetrics returns the singleton instance of operational metrics
func GetMetrics() *Metrics {
	return globalMetrics
}

// IncRequest increments total requests and per-agent counters
func (m *Metrics) IncRequest(agent string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests++
	m.RequestsByAgent[agent]++
}

// IncHandoff increments the silent handoff counter between agents
func (m *Metrics) IncHandoff() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalHandoffs++
}

// IncEscalate increments the human intervention counter
func (m *Metrics) IncEscalate() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalEscalates++
}
