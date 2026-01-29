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

func SetLLMClient(c llm.Client) {
	llmClient = c
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

var store = core.NewConversationStore(20)
var retriever *rag.Retriever

var brains = map[string]core.AgentBrain{
	"open_finance":      &openfinance.Brain{},
	"golpe_med":         &golpemed.Brain{},
	"atendimento_geral": &atendimento.Brain{},
	"criacao_conta":     &criacaoconta.Brain{},
}

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
		http.Error(w, "conversation_id and message are required", http.StatusBadRequest)
		return
	}

	// --- LOG DE ENTRADA ---
	log.Printf("trace=%s conv=%s event=request_received msg=\"%s\"", traceID, req.ConversationID, req.Message)

	// Registro inicial da mensagem do usuário
	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "user",
		Text:      req.Message,
		Timestamp: time.Now(),
	})

	var reply string
	var currentAction string = "reply"
	var finalAgent string

	for i := 0; i < 3; i++ {
		history := store.Get(req.ConversationID)

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

		ragText := ""
		if retriever != nil {
			ragText = retriever.AsText()
		}

		plan, err := brain.Run(r.Context(), llmClient, traceID, history, req.Message, ragText)
		if err != nil {
			// --- LOG DE ERRO NO BRAIN ---
			log.Printf("trace=%s conv=%s event=brain_error agent=%s err=%v", traceID, req.ConversationID, agent, err)
			reply = "Desculpe, tive um problema técnico. Pode repetir?"
			break
		}

		if plan.Action == "change_agent" {
			newAgent := plan.ChangeAgent
			if newAgent == "" || newAgent == "null" {
				newAgent = "atendimento_geral"
			}

			// --- LOG DE HANDOFF DETALHADO ---
			log.Printf("trace=%s conv=%s event=silent_handoff from=%s to=%s", traceID, req.ConversationID, agent, newAgent)

			store.SetAgent(req.ConversationID, newAgent)
			continue
		}

		currentAction = plan.Action
		reply = finalizeResponse(plan)
		break
	}

	// Registro da resposta final
	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "assistant",
		Text:      reply,
		Timestamp: time.Now(),
	})

	w.Header().Set("X-Trace-Id", traceID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(MessageResponse{
		Reply:        reply,
		Action:       currentAction,
		Agent:        finalAgent,
		HistoryCount: len(store.Get(req.ConversationID)),
		TraceID:      traceID,
	})

	log.Printf("trace=%s conv=%s event=replied agent=%s latency=%v", traceID, req.ConversationID, finalAgent, time.Since(start))
}

func finalizeResponse(plan core.ActionPlan) string {
	res := plan.Message
	if plan.Action == "ask" || plan.Action == "collect_data" {
		if plan.NextQuestion != "" {
			if res != "" {
				res += "\n\n"
			}
			res += plan.NextQuestion
		}
	}
	if res == "" {
		return "Como posso te ajudar com isso?"
	}
	return res
}
