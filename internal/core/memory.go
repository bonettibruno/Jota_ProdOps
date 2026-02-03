package core

import (
	"sync"
)

// ConversationStore manages in-memory chat history and agent state per conversation
type ConversationStore struct {
	mu     sync.Mutex
	items  map[string][]ChatMessage
	agents map[string]string
	limit  int
}

// NewConversationStore initializes a new store with a history limit per user
func NewConversationStore(limit int) *ConversationStore {
	return &ConversationStore{
		items:  make(map[string][]ChatMessage),
		agents: make(map[string]string),
		limit:  limit,
	}
}

// Add appends a new message to the conversation history and enforces the limit
func (s *ConversationStore) Add(convID string, msg ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := append(s.items[convID], msg)

	// Keep history within defined limit
	if s.limit > 0 && len(h) > s.limit {
		h = h[len(h)-s.limit:]
	}

	s.items[convID] = h
}

// Get retrieves a thread-safe copy of the conversation history
func (s *ConversationStore) Get(convID string) []ChatMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := s.items[convID]

	// Return a copy to prevent external mutation
	out := make([]ChatMessage, len(h))
	copy(out, h)
	return out
}

// GetAgent returns the current specialist agent assigned to the conversation
func (s *ConversationStore) GetAgent(convID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, ok := s.agents[convID]
	return agent, ok
}

// SetAgent updates the active specialist agent for a specific conversation
func (s *ConversationStore) SetAgent(convID, agent string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.agents[convID] = agent
}

// PrintAll dumps all active conversations to the console for debugging
func (s *ConversationStore) PrintAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	println("\n=== [MEMORY DUMP] ===")
	for id, messages := range s.items {
		agent := s.agents[id]
		if agent == "" {
			agent = "none"
		}

		println("ID:", id, "| AGENT:", agent, "| MSGS:", len(messages))
		for _, m := range messages {
			role := "U"
			if m.Role == "assistant" {
				role = "A"
			}

			text := m.Text
			if len(text) > 40 {
				text = text[:37] + "..."
			}
			println("  [" + role + "]: " + text)
		}
	}
	println("=====================\n")
}
