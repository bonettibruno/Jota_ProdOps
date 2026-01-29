package criacaoconta

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bonettibruno/Jota_ProdOps/internal/core"
	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
)

type Brain struct{}

// Run executes the Onboarding Specialist (Account Creation) agent logic
func (b *Brain) Run(ctx context.Context, client any, traceID string, history []core.ChatMessage, userMessage string, ragContext string) (core.ActionPlan, error) {

	// Type Assertion: retrieve the specific LLM client interface
	llmClient := client.(llm.Client)

	system := fmt.Sprintf(`Você é o Especialista em Onboarding (Criação de Conta) do Jota.
Sua identidade técnica é: "criacao_conta".

MISSÃO: Resolver problemas de selfie, documentos, erros de CPF/CNPJ e fluxos PF/PJ.

REGRAS DE OURO:
1. Se o problema for técnico (câmera/app), sugira troubleshoot (limpar cache, atualizar app).
2. Se o cliente enviar fotos aqui, diga que o envio é EXCLUSIVO pelo App do Jota.
3. Se o cliente mudar de assunto para algo que NÃO seja cadastro, você DEVE transferir.

NOMES PARA TRANSFERÊNCIA (campo change_agent):
- "open_finance": Se o cliente quiser conectar outros bancos.
- "golpe_med": Se o cliente relatar fraude, roubo ou golpe Pix.
- "atendimento_geral": Para dúvidas gerais não relacionadas a cadastro.

RESPOSTA EXCLUSIVAMENTE EM JSON:
{
  "action": "reply | ask | change_agent",
  "message": "sua resposta aqui",
  "change_agent": "nome_do_agente_alvo | null",
  "confidence": 1.0
}

Base de conhecimento:
%s`, ragContext)

	// Execute LLM text generation
	raw, err := llmClient.GenerateText(ctx, traceID, system, userMessage)
	if err != nil {
		return core.ActionPlan{}, err
	}

	var plan core.ActionPlan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return core.ActionPlan{}, fmt.Errorf("failed to unmarshal ActionPlan: %w", err)
	}

	return plan, nil
}
