package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/bonettibruno/Jota_ProdOps/internal/agents/atendimento"
	"github.com/bonettibruno/Jota_ProdOps/internal/agents/criacaoconta"
	"github.com/bonettibruno/Jota_ProdOps/internal/agents/golpemed"
	"github.com/bonettibruno/Jota_ProdOps/internal/agents/openfinance"
	"github.com/bonettibruno/Jota_ProdOps/internal/core"
	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
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

var llmClient llm.Client
var store = core.NewConversationStore(20)
var retriever *rag.Retriever

// Agent mapping for the Orchestrator
var brains = map[string]core.AgentBrain{
	"open_finance":      &openfinance.Brain{},
	"golpe_med":         &golpemed.Brain{},
	"atendimento_geral": &atendimento.Brain{},
	"criacao_conta":     &criacaoconta.Brain{},
}

func SetLLMClient(c llm.Client) {
	llmClient = c
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(core.GetMetrics())
}

func init() {
	// Initialize RAG retriever with knowledge base
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
	m := core.GetMetrics()

	// 1. Traceability: Unique ID for request tracking
	traceID := r.Header.Get("X-Trace-Id")
	if traceID == "" {
		traceID = newTraceID()
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Request validation
	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("trace=%s event=bad_json error=%v", traceID, err)
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.ConversationID == "" || req.Message == "" {
		http.Error(w, "conversation_id and message are required", http.StatusBadRequest)
		return
	}

	log.Printf("trace=%s conv=%s event=request_received msg=\"%s\"", traceID, req.ConversationID, req.Message)

	// 3. Persist user input in history
	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "user",
		Text:      req.Message,
		Timestamp: time.Now(),
	})

	var reply string
	var currentAction string = "reply"
	var finalAgent string

	// 4. Contextual RAG search
	ragText := ""
	if retriever != nil {
		ragText = retriever.SearchAsText(req.Message, 3)
		if ragText != "" {
			log.Printf("trace=%s conv=%s event=rag_retrieval status=success", traceID, req.ConversationID)
		}
	}

	// 5. Orchestration Loop (Policy Engine / Silent Handoff)
	for i := 0; i < 3; i++ {
		history := store.Get(req.ConversationID)
		log.Printf("trace=%s conv=%s event=debug_history messages_in_context=%d",
			traceID, req.ConversationID, len(history))

		agent, ok := store.GetAgent(req.ConversationID)
		if !ok {
			agent = "atendimento_geral"
			store.SetAgent(req.ConversationID, agent)
		}
		finalAgent = agent

		brain, exists := brains[agent]
		if !exists || llmClient == nil {
			reply = "Olá! Eu sou a Aline do Jota. Como posso te ajudar hoje?"
			break
		}

		// Execute specialized Agent Brain
		plan, err := brain.Run(r.Context(), llmClient, traceID, history, req.Message, ragText)
		if err != nil {
			log.Printf("trace=%s conv=%s event=brain_error agent=%s err=%v", traceID, req.ConversationID, agent, err)
			reply = "Desculpe, tive um problema técnico momentâneo. Pode repetir, por favor?"
			break
		}

		// Handle agent transition (Handoff)
		if plan.Action == "change_agent" {
			m.IncHandoff()
			newAgent := plan.ChangeAgent
			if newAgent == "" || newAgent == "null" {
				newAgent = "atendimento_geral"
			}

			log.Printf("trace=%s conv=%s event=silent_handoff from=%s to=%s reason=\"%s\"",
				traceID, req.ConversationID, agent, newAgent, plan.HandoffReason)

			store.SetAgent(req.ConversationID, newAgent)
			continue // Re-process with the new specialist
		}

		currentAction = plan.Action
		reply = finalizeResponse(plan)
		break
	}

	// 6. Persist final assistant response
	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "assistant",
		Text:      reply,
		Timestamp: time.Now(),
	})

	// Handle critical human intervention
	if currentAction == "escalate" {
		m.IncEscalate()
		log.Printf("trace=%s conv=%s event=HUMAN_INTERVENTION_REQUIRED level=CRITICAL agent=%s",
			traceID, req.ConversationID, finalAgent)
	}

	// Handle automated banking tools (e.g., MED)
	if currentAction == "call_api" {
		log.Printf("trace=%s conv=%s event=EXECUTING_TOOL tool=abrir_med status=simulated", traceID, req.ConversationID)
		reply = "Iniciei o protocolo MED (Mecanismo Especial de Devolução) para você. Vou acompanhar o processo e te aviso por aqui."
	}

	// 7. Send final response
	w.Header().Set("X-Trace-Id", traceID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(MessageResponse{
		Reply:        reply,
		Action:       currentAction,
		Agent:        finalAgent,
		HistoryCount: len(store.Get(req.ConversationID)),
		TraceID:      traceID,
	})

	store.PrintAll()

	m.IncRequest(finalAgent)
	log.Printf("trace=%s conv=%s event=replied agent=%s action=%s latency=%v",
		traceID, req.ConversationID, finalAgent, currentAction, time.Since(start))
}

func finalizeResponse(plan core.ActionPlan) string {
	res := plan.Message

	// Format specific action types
	switch plan.Action {
	case "call_api":
		if res == "" {
			res = "Entendido. Estou acionando nossos sistemas de segurança agora."
		}
		res += "\n\n[Protocolo de Segurança Ativado]"
	case "ask", "collect_data":
		if plan.NextQuestion != "" {
			if res != "" {
				res += "\n\n"
			}
			res += plan.NextQuestion
		}
	case "escalate":
		res = "Sinto muito por isso. " + res + "\n\nEstou transferindo você agora para um especialista humano. Por favor, aguarde."
	}

	if res == "" {
		return "Como posso te ajudar com isso?"
	}
	return res
}
