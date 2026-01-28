package core

import "strings"

func ShouldUseRAG(agent string, message string) bool {
	m := strings.ToLower(message)

	switch agent {
	case "open_finance":
		if strings.Contains(m, "como funciona") && strings.Contains(m, "open finance") {
			return true
		}
		if strings.Contains(m, "quais bancos") || strings.Contains(m, "bancos posso conectar") {
			return true
		}
		if strings.Contains(m, "ver o extrato") || (strings.Contains(m, "extrato") && strings.Contains(m, "conect")) {
			return true
		}
		return false

	case "atendimento_geral":
		if (strings.Contains(m, "limite") && strings.Contains(m, "pix")) || strings.Contains(m, "sem senha") {
			return true
		}
		if (strings.Contains(m, "cartão") || strings.Contains(m, "cartao")) &&
			(strings.Contains(m, "crédito") || strings.Contains(m, "credito")) {
			return true
		}
		if strings.Contains(m, "como funciona o jota") || strings.Contains(m, "como funciona") && strings.Contains(m, "jota") {
			return true
		}
		if strings.Contains(m, "horário de atendimento") || strings.Contains(m, "horario de atendimento") {
			return true
		}
		return false

	case "criacao_conta":
		return false

	case "golpe_med":
		return false

	default:
		return false
	}
}
