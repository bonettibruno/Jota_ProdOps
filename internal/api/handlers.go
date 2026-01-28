package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bonettibruno/Jota_ProdOps/internal/core"
)

type MessageRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

type MessageResponse struct {
	Reply        string `json:"reply"`
	Action       string `json:"action"`
	Agent        string `json:"agent"`
	HistoryCount int    `json:"history_count"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

var store = core.NewConversationStore(20)

func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.ConversationID == "" {
		http.Error(w, "conversation_id is required", http.StatusBadRequest)
		return
	}

	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "user",
		Text:      req.Message,
		Timestamp: time.Now(),
	})

	history := store.Get(req.ConversationID)

	agent, ok := store.GetAgent(req.ConversationID)
	if !ok {
		agent = core.RouteAgent(req.Message)
		store.SetAgent(req.ConversationID, agent)
	}

	reply := core.GenerateReply(agent, history)

	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "assistant",
		Text:      reply,
		Timestamp: time.Now(),
	})

	resp := MessageResponse{
		Reply:        reply,
		Action:       "reply",
		Agent:        agent,
		HistoryCount: len(history) + 1,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)

}
