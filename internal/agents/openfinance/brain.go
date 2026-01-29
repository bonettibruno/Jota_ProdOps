package openfinance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bonettibruno/Jota_ProdOps/internal/core"
	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
)

type Brain struct{}

// Run executes the Open Finance specialist agent logic
func (b *Brain) Run(
	ctx context.Context,
	client any,
	traceID string,
	history []core.ChatMessage,
	userMessage string,
	ragContext string,
) (core.ActionPlan, error) {

	// Type Assertion: retrieve the specific LLM client interface
	llmClient := client.(llm.Client)

	systemPrompt := buildSystemPrompt(ragContext)
	userPrompt := buildUserPrompt(history, userMessage)

	// Execute LLM text generation
	raw, err := llmClient.GenerateText(ctx, traceID, systemPrompt, userPrompt)
	if err != nil {
		return core.ActionPlan{}, err
	}

	var plan core.ActionPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return core.ActionPlan{}, fmt.Errorf("failed to decode ActionPlan JSON: %w", err)
	}

	return plan, nil
}

// buildSystemPrompt defines behavioral guidelines and RAG context
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

// buildUserPrompt formats history and current message for the model
func buildUserPrompt(history []core.ChatMessage, userMessage string) string {
	h := ""
	for _, msg := range history {
		role := "Cliente"
		if msg.Role == "assistant" {
			role = "Você (Especialista)"
		}
		h += fmt.Sprintf("%s: %s\n", role, msg.Text)
	}

	return fmt.Sprintf(`Histórico da conversa:
%s

Mensagem atual do cliente:
"%s"

Gere o ActionPlan em JSON:`, h, userMessage)
}
