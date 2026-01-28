package core

import "strings"

func RouteAgent(message string) string {
	m := strings.ToLower(message)

	if strings.Contains(m, "golpe") ||
		strings.Contains(m, "fraude") ||
		strings.Contains(m, "pix") ||
		strings.Contains(m, "invad") ||
		strings.Contains(m, "hack") {
		return "golpe_med"
	}

	if strings.Contains(m, "open finance") ||
		strings.Contains(m, "conectar") ||
		strings.Contains(m, "vincular") ||
		strings.Contains(m, "itau") || strings.Contains(m, "itaú") ||
		strings.Contains(m, "nubank") ||
		strings.Contains(m, "inter") ||
		strings.Contains(m, "bradesco") ||
		strings.Contains(m, "santander") {
		return "open_finance"
	}

	if strings.Contains(m, "abrir conta") ||
		strings.Contains(m, "criar conta") ||
		strings.Contains(m, "cadastro") ||
		strings.Contains(m, "selfie") ||
		strings.Contains(m, "documento") ||
		strings.Contains(m, "cpf") ||
		strings.Contains(m, "cnpj") {
		return "criacao_conta"
	}

	return "atendimento_geral"
}
