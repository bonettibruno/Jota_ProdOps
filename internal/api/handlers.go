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

	// 1. Setup inicial e Trace ID
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

	// 3. Registro da mensagem e busca de histórico
	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "user",
		Text:      req.Message,
		Timestamp: time.Now(),
	})
	history := store.Get(req.ConversationID)

	var historyTexts []string
	for _, h := range history {
		historyTexts = append(historyTexts, h.Role+": "+h.Text)
	}

	// 4. Lógica de Roteamento
	agent, ok := store.GetAgent(req.ConversationID)
	if !ok {
		agent = "atendimento_geral"
		if llmClient != nil {
			decision, err := llmClient.RouteAgent(r.Context(), traceID, req.Message, historyTexts)
			if err == nil && decision.Agent != "" {
				agent = decision.Agent
				store.SetAgent(req.ConversationID, agent)
				log.Printf("trace=%s event=agent_routed decision=%s", traceID, agent)
			} else if err != nil {
				log.Printf("trace=%s event=routing_error error=%v", traceID, err)
			}
		}
	}

	var citations []core.Citation
	var reply string
	var currentAction string = "reply"

	// 5. Execução do Agent Brain
	// Certifique-se de que 'agent' aqui vale exatamente "golpe_med"
	brain, exists := brains[agent]

	if exists && llmClient != nil {
		ragText := ""
		if retriever != nil {
			ragText = retriever.AsText()
		}
		plan, err := brain.Run(r.Context(), llmClient, traceID, history, req.Message, ragText)
		if err != nil {
			log.Printf("trace=%s event=brain_error agent=%s err=%v", traceID, agent, err)
			reply = "Desculpe, tive um problema técnico para processar sua solicitação. Pode repetir, por favor?"
		} else {
			currentAction = plan.Action
			reply = processActionPlan(req.ConversationID, &agent, plan)
		}
	} else {
		log.Printf("trace=%s event=using_legacy_reply agent=%s", traceID, agent)
		reply = "Olá! Eu sou a Aline do Jota. Como posso te ajudar hoje?"
	}

	// 6. Finalização
	store.Add(req.ConversationID, core.ChatMessage{
		Role:      "assistant",
		Text:      reply,
		Timestamp: time.Now(),
	})

	w.Header().Set("X-Trace-Id", traceID)
	w.Header().Set("Content-Type", "application/json")
	resp := MessageResponse{
		Reply:        reply,
		Action:       currentAction,
		Agent:        agent,
		HistoryCount: len(history) + 1,
		TraceID:      traceID,
		Citations:    citations,
	}
	_ = json.NewEncoder(w).Encode(resp)

	log.Printf("trace=%s event=replied agent=%s latency=%v", traceID, agent, time.Since(start))
}

func processActionPlan(convID string, currentAgent *string, plan core.ActionPlan) string {
	// 1. Prioridade: Se o agente decidiu trocar, atualizamos o estado
	if plan.Action == "change_agent" {
		newAgent := plan.ChangeAgent
		if newAgent == "" || newAgent == "null" {
			newAgent = "atendimento_geral" // Fallback de segurança
		}

		log.Printf("event=executing_handoff conversation_id=%s from=%s to=%s", convID, *currentAgent, newAgent)

		store.SetAgent(convID, newAgent)
		*currentAgent = newAgent

		// Se a LLM não mandou mensagem de transferência, nós geramos uma
		if plan.Message == "" {
			return "Entendido! Vou te transferir para o especialista em " + newAgent + " para continuarmos."
		}
		return plan.Message
	}

	// 2. Tratamento para mensagens de coleta de dados ou perguntas
	if plan.Action == "ask" || plan.Action == "collect_data" {
		res := plan.Message
		if plan.NextQuestion != "" {
			if res != "" {
				res += "\n\n"
			}
			res += plan.NextQuestion
		}
		if res == "" {
			return "Pode me dar mais detalhes sobre isso para que eu possa te ajudar?"
		}
		return res
	}

	// 3. Fallback: Se nada acima pegou mas tem mensagem, manda a mensagem
	if plan.Message != "" {
		return plan.Message
	}

	// 4. Último recurso para evitar o JSON {"reply": ""}
	return "Entendi. Como posso te ajudar com isso?"
}
