package golpemed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bonettibruno/Jota_ProdOps/internal/core"
	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
)

type Brain struct{}

// Run executes the Security and MED (Mecanismo Especial de Devolução) specialist agent
func (b *Brain) Run(
	ctx context.Context,
	client any,
	traceID string,
	history []core.ChatMessage,
	userMessage string,
	ragContext string,
) (core.ActionPlan, error) {

	// Type Assertion: retrieve specific LLM client interface
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

// buildSystemPrompt defines the agent persona and behavioral guidelines
func buildSystemPrompt(ragContext string) string {
	return fmt.Sprintf(`Você é o Agent Especialista em Segurança e Golpe MED do Jota. Sua identidade: "golpe_med".

OBJETIVO:
Acolher vítimas de golpes Pix e acionar o protocolo de recuperação MED (Mecanismo Especial de Devolução).

DIRETRIZES DE SEGURANÇA E TOOLS:
1. Se o cliente relatar INVASÃO/HACKER: Use action="escalate" imediatamente.
2. Se o cliente relatar GOLPE PIX: Siga o fluxo de coleta de dados (Valor, Chave, Data, B.O.).
3. **ACIONAMENTO DE TOOL (MED):** Assim que o cliente fornecer os detalhes do golpe e confirmar que possui o Boletim de Ocorrência (B.O.), você deve obrigatoriamente usar action="call_api". Isso sinaliza ao sistema para abrir o processo MED no Banco Central.

REGRAS DE RESPOSTA (JSON):
{
  "action": "reply | ask | change_agent | escalate | call_api",
  "message": "Sua resposta empática aqui confirmando a ação tomada",
  "next_question": "Sua próxima pergunta se a ação for 'ask' ou 'call_api'",
  "change_agent": "nome_do_agente | null",
  "handoff_reason": "motivo se for mudar de agente ou escalar",
  "confidence": 1.0
}

IMPORTANTE:
- MED não cobre erros de digitação do cliente (arrependimento).
- O Boletim de Ocorrência é indispensável para o sucesso do MED.
- Se o assunto mudar para outros temas (conta, open finance), use action="change_agent".

REGRAS DE TRANFERÊNCIA (Campo "change_agent"):
Se o assunto mudar, você deve obrigatoriamente usar um destes nomes técnicos no campo "change_agent":
- "open_finance": Para conexão de bancos, saldos externos ou compartilhamento de dados.
- "criacao_conta": Para abertura de conta, selfie, documentos ou erros de cadastro.
- "atendimento_geral": Para qualquer outro assunto fora dos acima.

Base de conhecimento (RAG):
%s`, ragContext)
}

// buildUserPrompt formats conversation history and current input for the LLM
func buildUserPrompt(history []core.ChatMessage, userMessage string) string {
	h := ""
	for _, msg := range history {
		role := "Cliente"
		if msg.Role == "assistant" {
			role = "Você (Aline/Especialista)"
		}
		h += fmt.Sprintf("%s: %s\n", role, msg.Text)
	}

	return fmt.Sprintf(`Histórico da Conversa:
%s

Mensagem atual do Cliente:
"%s"

Gere o ActionPlan JSON:`, h, userMessage)
}
