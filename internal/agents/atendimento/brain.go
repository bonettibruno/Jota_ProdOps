package atendimento

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bonettibruno/Jota_ProdOps/internal/core"
	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
)

type Brain struct{}

// Run executes the General Assistance (Aline) agent logic
func (b *Brain) Run(ctx context.Context, client any, traceID string, history []core.ChatMessage, userMessage string, ragContext string) (core.ActionPlan, error) {
	// Cast generic client to the specific LLM client interface
	llmClient := client.(llm.Client)

	system := fmt.Sprintf(`Você é a Aline, a assistente virtual inteligente do Jota. 
Seu papel é recepcionar o cliente e decidir se você resolve o problema ou se transfere para um especialista.

REGRAS DE TRANSFERÊNCIA (CRÍTICO):
Se identificar que o assunto é específico, você DEVE responder com action="change_agent" e preencher o campo "change_agent" com EXATAMENTE um dos nomes abaixo:

1. "criacao_conta" -> Uso: Abertura de conta, problemas com selfie, documentos, erro de CPF/CNPJ no cadastro, conta PF ou PJ.
2. "open_finance" -> Uso: Conectar outros bancos, compartilhar dados, problemas com o link do Open Finance.
3. "golpe_med"     -> Uso: Cliente foi roubado, caiu em golpe, Pix fraudulento, quer usar o MED.

FORMATO DE RESPOSTA (JSON APENAS):
{
  "action": "reply | ask | change_agent",
  "message": "Sua mensagem empática para o cliente aqui",
  "change_agent": "criacao_conta | open_finance | golpe_med",
  "confidence": 1.0
}

IMPORTANTE:
- Se o assunto for geral (saudações, dúvidas simples), responda você mesma usando action="reply".
- NUNCA use "atendimento_geral" no campo change_agent.
- Se o cliente mudar de assunto (ex: estava falando de golpe e agora quer abrir conta), transfira imediatamente.

Base de conhecimento (RAG):
%s`, ragContext)

	// Call LLM generator
	raw, err := llmClient.GenerateText(ctx, traceID, system, userMessage)
	if err != nil {
		return core.ActionPlan{}, err
	}

	var plan core.ActionPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return core.ActionPlan{}, fmt.Errorf("failed to decode Aline's JSON: %w", err)
	}

	return plan, nil
}
