package llm

import "context"

type RouterDecision struct {
	Agent      string  `json:"agent"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

type Client interface {
	RouteAgent(ctx context.Context, traceID string, message string, history []string) (RouterDecision, error)
}
