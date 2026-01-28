package core

func AgentReply(agent string) string {
	switch agent {
	case "golpe_med":
		return "Olá! Sinto muito pelo ocorrido. Para te ajudar, preciso de algumas informações: qual foi o valor, a data/hora e para qual chave/conta você enviou o Pix?"
	case "open_finance":
		return "Entendi! Vamos resolver. Primeiro, copie o link e tente abrir no Chrome ou Safari (fora do WhatsApp). Se aparecer erro, me diga qual mensagem aparece."
	case "criacao_conta":
		return "Claro! Você está tentando abrir conta ou está com erro no cadastro (ex.: selfie/documento)? Me diga exatamente o que aparece na tela para eu te orientar."
	default:
		return "Olá! Eu sou a Aline, assistente virtual do Jota. Para eu te direcionar melhor, você pode me dizer em uma frase qual é a sua necessidade?"
	}
}
