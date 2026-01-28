package api

import (
	"encoding/json"
	"net/http"

	"github.com/bonettibruno/Jota_ProdOps/internal/core"
)

type MessageRequest struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Reply  string `json:"reply"`
	Action string `json:"action"`
	Agent  string `json:"agent"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

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

	agent := core.RouteAgent(req.Message)
	reply := core.AgentReply(agent)

	resp := MessageResponse{
		Reply:  reply,
		Action: "reply",
		Agent:  agent,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
