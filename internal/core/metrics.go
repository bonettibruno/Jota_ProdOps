package core

import "sync"

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

func GetMetrics() *Metrics {
	return globalMetrics
}

func (m *Metrics) IncRequest(agent string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalRequests++
	m.RequestsByAgent[agent]++
}

func (m *Metrics) IncHandoff() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalHandoffs++
}

func (m *Metrics) IncEscalate() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.TotalEscalates++
}
