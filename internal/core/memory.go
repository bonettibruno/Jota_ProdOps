package core

import (
	"sync"
	"time"
)

type ChatMessage struct {
	Role      string // "user" ou "assistant"
	Text      string
	Timestamp time.Time
}

type ConversationStore struct {
	mu     sync.Mutex
	items  map[string][]ChatMessage
	agents map[string]string
	limit  int
}

func NewConversationStore(limit int) *ConversationStore {
	return &ConversationStore{
		items:  make(map[string][]ChatMessage),
		agents: make(map[string]string),
		limit:  limit,
	}
}

func (s *ConversationStore) Add(convID string, msg ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := append(s.items[convID], msg)

	if s.limit > 0 && len(h) > s.limit {
		h = h[len(h)-s.limit:]
	}

	s.items[convID] = h
}

func (s *ConversationStore) Get(convID string) []ChatMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	h := s.items[convID]

	out := make([]ChatMessage, len(h))
	copy(out, h)
	return out
}

func (s *ConversationStore) GetAgent(convID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	agent, ok := s.agents[convID]
	return agent, ok
}

func (s *ConversationStore) SetAgent(convID, agent string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.agents[convID] = agent
}
