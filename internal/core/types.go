package core

import (
	"context"
	"time"
)

// ChatMessage represents a single interaction in the conversation history
type ChatMessage struct {
	Role      string    `json:"role"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// ActionPlan defines the schema for LLM decision-making and routing
type ActionPlan struct {
	Action        string  `json:"action"`
	Message       string  `json:"message"`
	NextQuestion  string  `json:"next_question"`
	ChangeAgent   string  `json:"change_agent"`
	HandoffReason string  `json:"handoff_reason"`
	Confidence    float64 `json:"confidence"`
}

// AgentBrain defines the interface for specialized agent logic
type AgentBrain interface {
	// Run executes the agent's logic and returns an ActionPlan
	Run(ctx context.Context, client any, traceID string, history []ChatMessage, userMessage string, ragContext string) (ActionPlan, error)
}

// Citation represents a reference from the Knowledge Base (RAG)
type Citation struct {
	Source  string `json:"source"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}
