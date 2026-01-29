package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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

	// 2. Decode da Request (o que faz o 'req' parar de ficar vermelho)
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
			if err == nil {
				agent = decision.Agent
				store.SetAgent(req.ConversationID, agent)
			}
		}
	}

	var citations []core.Citation
	var reply string
	var currentAction string = "reply"

	// 5. Execução do Agent Brain
	if agent == "open_finance" && llmClient != nil {
		ragText := ""
		if retriever != nil {
			ragText = retriever.AsText() // Certifique-se que este método existe no seu rag/simple.go
		}

		plan, err := openfinance.RunBrain(r.Context(), llmClient, traceID, history, req.Message, ragText)
		if err != nil {
			log.Printf("trace=%s event=agent_brain_failed error=%v", traceID, err)
			reply = core.GenerateReply(agent, history)
		} else {
			currentAction = plan.Action
			switch plan.Action {
			case "reply", "ask":
				reply = plan.Message
				if plan.NextQuestion != "" {
					reply += "\n\n" + plan.NextQuestion
				}
			case "change_agent":
				if plan.ChangeAgent != "" {
					store.SetAgent(req.ConversationID, plan.ChangeAgent)
					reply = "Entendi. Vou te transferir para o especialista em " + plan.ChangeAgent + "."
					agent = plan.ChangeAgent
				}
			case "escalate":
				reply = plan.Message
			default:
				reply = plan.Message
			}
		}
	} else {
		if retriever != nil && core.ShouldUseRAG(agent, req.Message) {
			reply, citations = core.BuildReplyWithRAG(agent, req.Message, retriever)
		} else {
			reply = core.GenerateReply(agent, history)
		}
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
