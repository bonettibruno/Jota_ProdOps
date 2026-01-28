package core

import "strings"

func GenerateReply(agent string, history []ChatMessage) string {
	// Pega a última mensagem do usuário (a mais recente)
	lastUser := ""
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "user" {
			lastUser = history[i].Text
			break
		}
	}
	m := strings.ToLower(lastUser)

	switch agent {
	case "golpe_med":
		if strings.Contains(m, "r$") || strings.Contains(m, "reais") {
			return "Obrigada! Agora me diga a data/horário da transação e para qual chave Pix/conta você enviou. Se tiver, envie também prints da conversa/comprovante."
		}
		return "Olá! Sinto muito pelo ocorrido. Para te ajudar, me diga: qual foi o valor, a data/hora e para qual chave Pix/conta você enviou?"

	case "open_finance":
		return "Entendi! Vamos resolver. Primeiro, copie o link e tente abrir no Chrome/Safari (fora do WhatsApp). Se aparecer erro, me diga qual mensagem aparece e qual banco você está tentando conectar."

	case "criacao_conta":
		return "Claro! Você está tentando abrir conta ou está com erro no cadastro (ex.: selfie/documento)? Me diga exatamente o que aparece na tela para eu te orientar."

	default:
		return "Olá! Eu sou a Aline, assistente virtual do Jota. Para eu te direcionar melhor, me diga em uma frase qual é a sua necessidade."
	}
}
