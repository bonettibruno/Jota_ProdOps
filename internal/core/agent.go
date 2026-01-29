package core

import (
	"context"

	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
)

type ActionPlan struct {
	Action        string  `json:"action"`
	Message       string  `json:"message,omitempty"`
	NextQuestion  string  `json:"next_question,omitempty"`
	ChangeAgent   string  `json:"change_agent,omitempty"`
	HandoffReason string  `json:"handoff_reason,omitempty"`
	Confidence    float64 `json:"confidence"`
}

type AgentBrain interface {
	Run(ctx context.Context, client llm.Client, traceID string, history []ChatMessage, userMessage string, ragContext string) (ActionPlan, error)
}
