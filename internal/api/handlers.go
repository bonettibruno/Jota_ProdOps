package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bonettibruno/Jota_ProdOps/internal/core"
	"github.com/bonettibruno/Jota_ProdOps/internal/rag"
)

type MessageRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
}

type MessageResponse struct {
	Reply        string          `json:"reply"`
	Action       string          `json:"action"`
	Agent        string          `json:"agent"`
	HistoryCount int             `json:"history_count"`
	TraceID      string          `json:"trace_id"`
	Citations    []core.Citation `json:"citations,omitempty"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

var store = core.NewConversationStore(20)

var retriever *rag.Retriever

func init() {
	r, err := rag.NewRetriever("kb/RAG_JOTA_RESUMIDO.md")
	if err != nil {
		log.Printf("event=rag_init_failed error=%v", err)
		return
	}
	retriever = r
	log.Printf("event=rag_ready source=kb/RAG_JOTA_RESUMIDO.md")
}

func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	traceID := r.Header.Get("X-Trace-Id")
	if traceID == "" {
		traceID = newTraceID()
	}

	if r.Method != http.MethodPost {
		log.Printf("trace=%s event=method_not_allowed method=%s path=%s", traceID, r.Method, r.URL.Path)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("trace=%s event=bad_json error=%v", traceID, err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.ConversationID == "" || req.Message == "" {
		log.Printf("trace=%s event=bad_request conversation_id=%q message_empty=%t", traceID, req.ConversationID, req.Message == "")
		http.Error(w, "conversation_id and message are required", http.StatusBadRequest)
		return
	}

	log.Printf("trace=%s event=message_received conversation_id=%s msg=%q", traceID, req.ConversationID, req.Message)

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
		log.Printf("trace=%s event=agent_routed conversation_id=%s agent=%s sticky=false", traceID, req.ConversationID, agent)
	} else {
		log.Printf("trace=%s event=agent_sticky conversation_id=%s agent=%s sticky=true", traceID, req.ConversationID, agent)
	}

	var citations []core.Citation
	var reply string

	if retriever != nil {
		reply, citations = core.BuildReplyWithRAG(agent, req.Message, retriever)
	} else {
		reply = core.GenerateReply(agent, history)
	}

	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "assistant",
		Text:      reply,
		Timestamp: time.Now(),
	})

	w.Header().Set("X-Trace-Id", traceID)
	w.Header().Set("Content-Type", "application/json")

	resp := MessageResponse{
		Reply:        reply,
		Action:       "reply",
		Agent:        agent,
		HistoryCount: len(history) + 1,
		TraceID:      traceID,
		Citations:    citations,
	}

	_ = json.NewEncoder(w).Encode(resp)

	log.Printf(
		"trace=%s event=message_replied conversation_id=%s agent=%s history_count=%d latency_ms=%d",
		traceID,
		req.ConversationID,
		agent,
		len(history)+1,
		time.Since(start).Milliseconds(),
	)
}
