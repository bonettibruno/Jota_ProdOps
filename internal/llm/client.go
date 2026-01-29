package llm

import "context"

// RouterDecision represents the output of the initial classification step
type RouterDecision struct {
	Agent      string  `json:"agent"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// Client defines the contract for Large Language Model integrations
type Client interface {
	// RouteAgent determines the most suitable specialized agent for a given input
	RouteAgent(
		ctx context.Context,
		traceID string,
		message string,
		history []string,
	) (RouterDecision, error)

	// GenerateText handles generic text generation with system instruction support
	GenerateText(
		ctx context.Context,
		traceID string,
		systemPrompt string,
		userPrompt string,
	) (string, error)
}
