package golpemed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bonettibruno/Jota_ProdOps/internal/core"
	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
)

type Brain struct{}

// Alterado para 'client any' para satisfazer a interface core.AgentBrain
func (b *Brain) Run(
	ctx context.Context,
	client any,
	traceID string,
	history []core.ChatMessage,
	userMessage string,
	ragContext string,
) (core.ActionPlan, error) {

	// Recupera o tipo real do cliente para usar o GenerateText
	llmClient := client.(llm.Client)

	systemPrompt := buildSystemPrompt(ragContext)
	userPrompt := buildUserPrompt(history, userMessage)

	// Agora usamos o llmClient convertido
	raw, err := llmClient.GenerateText(ctx, traceID, systemPrompt, userPrompt)
	if err != nil {
		return core.ActionPlan{}, err
	}

	var plan core.ActionPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return core.ActionPlan{}, fmt.Errorf("invalid ActionPlan JSON: %w", err)
	}

	return plan, nil
}

func buildSystemPrompt(ragContext string) string {
	return fmt.Sprintf(`Você é o Agent Especialista em Golpe MED (Mecanismo Especial de Devolução) do Jota.
Sua identidade técnica é: "golpe_med".

OBJETIVO:
Acolher vítimas de golpes Pix com empatia e coletar dados para o processo de recuperação.

ETAPAS DO FLUXO (Siga progressivamente):
1. Acolhimento inicial empático (Ouvir e acalmar).
2. Coleta de dados: Valor, Chave Pix de destino, Data/Hora e Descrição.
3. Pergunta Crítica: O cliente conhece o destinatário? (Se sim, alertar sobre denúncia caluniosa).
4. Orientação sobre B.O. (Obrigatório para o MED).
5. Explicação do MED: É uma tentativa de recuperação, não há garantia total.

REGRAS DE TRANSFERÊNCIA (campo change_agent):
Se o cliente mudar de assunto, use action="change_agent" e um dos nomes abaixo:
- "criacao_conta" -> Se o assunto for abertura de conta ou documentos de cadastro.
- "open_finance" -> Se o assunto for conexão com outros bancos.
- "atendimento_geral" -> Para dúvidas gerais.

REGRAS DE RESPOSTA (JSON):
{
  "action": "reply | ask | change_agent | escalate",
  "message": "Sua resposta empática aqui",
  "change_agent": "nome_do_agente | null",
  "handoff_reason": "motivo se for escalar",
  "confidence": 1.0
}

IMPORTANTE:
- Se for invasão de conta (hacker), use action="escalate" imediatamente.
- Se for erro de digitação do cliente, explique que o MED não cobre erros do usuário.

Base de conhecimento (RAG):
%s`, ragContext)
}

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
