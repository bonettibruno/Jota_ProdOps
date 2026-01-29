package openfinance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bonettibruno/Jota_ProdOps/internal/core"
	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
)

type ActionPlan struct {
	Action        string  `json:"action"`
	Message       string  `json:"message,omitempty"`
	NextQuestion  string  `json:"next_question,omitempty"`
	ChangeAgent   string  `json:"change_agent,omitempty"`
	HandoffReason string  `json:"handoff_reason,omitempty"`
	Confidence    float64 `json:"confidence"`
}

func RunBrain(
	ctx context.Context,
	client llm.Client,
	traceID string,
	history []core.ChatMessage,
	userMessage string,
	ragContext string,
) (ActionPlan, error) {

	systemPrompt := buildSystemPrompt(ragContext)
	userPrompt := buildUserPrompt(history, userMessage)

	// Agora enviamos System e User separados
	raw, err := client.GenerateText(ctx, traceID, systemPrompt, userPrompt)
	if err != nil {
		return ActionPlan{}, err
	}

	var plan ActionPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return ActionPlan{}, fmt.Errorf("invalid ActionPlan JSON: %w\nraw=%s", err, raw)
	}

	if plan.Action == "" {
		return ActionPlan{}, fmt.Errorf("action plan sem action")
	}

	return plan, nil
}

// buildSystemPrompt foca nas REGRAS e CONTEXTO RAG
func buildSystemPrompt(ragContext string) string {
	return fmt.Sprintf(`Você é o Agent de Open Finance do Jota.

Seu papel:
- Resolver problemas de conexão de Open Finance.
- Seguir fluxo progressivo: 1) link no navegador, 2) app visível, 3) titularidade PF/PJ, 4) trocar rede/navegador, 5) testar outro banco.
- Se houver frustração ou muitas tentativas, escalar para humano.

Regras:
- Responda EXCLUSIVAMENTE com JSON válido.
- Se faltar informação para ajudar, use action="ask".
- Se o assunto NÃO for Open Finance, use action="change_agent" e indique o agente correto.

Formato da resposta:
{
  "action": "reply | ask | change_agent | escalate | end",
  "message": "texto curto e direto para o cliente",
  "next_question": "pergunta para continuar o fluxo, se houver",
  "change_agent": "atendimento_geral | criacao_conta | golpe_med | null",
  "handoff_reason": "motivo da escalação ou troca",
  "confidence": 0.0
}

Base de conhecimento (RAG):
%s`, ragContext)
}

// buildUserPrompt foca na DINÂMICA da conversa atual
func buildUserPrompt(history []core.ChatMessage, userMessage string) string {
	h := ""
	for _, msg := range history {
		h += fmt.Sprintf("%s: %s\n", msg.Role, msg.Text)
	}

	return fmt.Sprintf(`Histórico da conversa:
%s

Mensagem atual do cliente:
"%s"

Gere o ActionPlan em JSON:`, h, userMessage)
}
