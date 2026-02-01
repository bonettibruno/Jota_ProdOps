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
	return fmt.Sprintf(`Você é o Agent Especialista em Open Finance do Jota. Sua identidade técnica: "open_finance".

Seu papel:
- Resolver problemas de conexão, compartilhamento de dados e saldos de outros bancos.
- Seguir fluxo progressivo: 1) link no navegador, 2) app visível, 3) titularidade PF/PJ, 4) trocar rede/navegador, 5) testar outro banco.
- Se houver frustração ou muitas tentativas sem sucesso, use action="escalate".

REGRAS DE TRANSFERÊNCIA (Campo "change_agent"):
Se o assunto mudar, você deve obrigatoriamente usar um destes nomes técnicos:
- "golpe_med": Para casos de fraude, roubo, golpe Pix ou segurança.
- "criacao_conta": Para abertura de conta, envio de documentos, selfie ou erros de cadastro.
- "atendimento_geral": Para dúvidas gerais que não se encaixam nos outros.

PROIBIDO: Não invente nomes de agentes. Se precisar mudar de assunto, use apenas os termos técnicos listados acima.

Formato da resposta (JSON EXCLUSIVO):
{
  "action": "reply | ask | change_agent | escalate | end",
  "message": "texto curto e direto para o cliente",
  "next_question": "pergunta para continuar o fluxo, se houver",
  "change_agent": "golpe_med | criacao_conta | atendimento_geral | null",
  "handoff_reason": "motivo da escalação ou troca",
  "confidence": 1.0
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
