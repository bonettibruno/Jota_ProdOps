package core

import (
	"context"
	"time"
)

// ChatMessage representa uma mensagem no histórico
type ChatMessage struct {
	Role      string    `json:"role"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// ActionPlan é o contrato de saída das LLMs
type ActionPlan struct {
	Action        string  `json:"action"`
	Message       string  `json:"message"`
	NextQuestion  string  `json:"next_question"`
	ChangeAgent   string  `json:"change_agent"`
	HandoffReason string  `json:"handoff_reason"`
	Confidence    float64 `json:"confidence"`
}

// AgentBrain simplificado para aceitar o cliente como 'any'
type AgentBrain interface {
	Run(ctx context.Context, client any, traceID string, history []ChatMessage, userMessage string, ragContext string) (ActionPlan, error)
}

// Citation representa uma fonte do RAG
type Citation struct {
	Source  string `json:"source"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}
