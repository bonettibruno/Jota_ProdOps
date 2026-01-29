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

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(core.GetMetrics())
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

	m := core.GetMetrics()

	// 1. Rastreabilidade: Garantir que toda requisição tenha um ID único
	traceID := r.Header.Get("X-Trace-Id")
	if traceID == "" {
		traceID = newTraceID()
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Ingestão e Validação
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

	// 3. Persistência da entrada do usuário
	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "user",
		Text:      req.Message,
		Timestamp: time.Now(),
	})

	var reply string
	var currentAction string = "reply"
	var finalAgent string

	// --- OTIMIZAÇÃO RAG: Busca contextual única para economizar tokens e latência ---
	ragText := ""
	if retriever != nil {
		// Buscamos os 3 trechos mais relevantes da base de conhecimento
		ragText = retriever.SearchAsText(req.Message, 3)
		if ragText != "" {
			log.Printf("trace=%s conv=%s event=rag_retrieval status=success", traceID, req.ConversationID)
		}
	}

	// 4. Loop de Orquestração (Policy Engine / Handoff Silencioso)
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

		// Execução do "Cérebro" do Agente atual
		plan, err := brain.Run(r.Context(), llmClient, traceID, history, req.Message, ragText)
		if err != nil {
			log.Printf("trace=%s conv=%s event=brain_error agent=%s err=%v", traceID, req.ConversationID, agent, err)
			reply = "Desculpe, tive um problema técnico momentâneo. Pode repetir, por favor?"
			break
		}

		// Verifica se o Agente solicitou troca de especialista (Handoff)
		if plan.Action == "change_agent" {
			m.IncHandoff()
			newAgent := plan.ChangeAgent
			if newAgent == "" || newAgent == "null" {
				newAgent = "atendimento_geral"
			}

			log.Printf("trace=%s conv=%s event=silent_handoff from=%s to=%s reason=\"%s\"",
				traceID, req.ConversationID, agent, newAgent, plan.HandoffReason)

			store.SetAgent(req.ConversationID, newAgent)
			continue // Re-processa com o novo agente sem responder ao usuário ainda
		}

		// Se chegou aqui, o agente decidiu responder ou agir
		currentAction = plan.Action
		reply = finalizeResponse(plan)
		break
	}

	// 5. Registro da resposta final na memória
	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "assistant",
		Text:      reply,
		Timestamp: time.Now(),
	})

	// Este sinal indica que a IA identificou uma situação que exige um especialista humano
	if currentAction == "escalate" {
		m.IncEscalate()
		fullHistory := store.Get(req.ConversationID)

		log.Printf("trace=%s conv=%s event=HUMAN_INTERVENTION_REQUIRED level=CRITICAL agent=%s reason=\"Ação de risco detectada\"",
			traceID, req.ConversationID, finalAgent)

		// Aqui você enviaria o 'fullHistory' para o seu sistema de Chat (ex: Intercom/Zendesk)
		// O humano já recebe a conversa pronta, sem precisar perguntar tudo de novo.
		_ = fullHistory
	}

	// 6. Resposta Final para o Canal (WhatsApp/Web)
	w.Header().Set("X-Trace-Id", traceID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(MessageResponse{
		Reply:        reply,
		Action:       currentAction,
		Agent:        finalAgent,
		HistoryCount: len(store.Get(req.ConversationID)),
		TraceID:      traceID,
	})

	m.IncRequest(finalAgent)
	// Log de Observabilidade: Latência é a métrica principal aqui
	log.Printf("trace=%s conv=%s event=replied agent=%s action=%s latency=%v",
		traceID, req.ConversationID, finalAgent, currentAction, time.Since(start))
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
