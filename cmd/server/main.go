package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type MessageRequest struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Reply  string `json:"reply"`
	Action string `json:"action"`
	Agent  string `json:"agent"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func routeAgent(message string) string {
	m := strings.ToLower(message)

	if strings.Contains(m, "golpe") || strings.Contains(m, "fraude") || strings.Contains(m, "pix") || strings.Contains(m, "invad") {
		return "golpe_med"
	}

	if strings.Contains(m, "open finance") || strings.Contains(m, "conectar") || strings.Contains(m, "vincular") ||
		strings.Contains(m, "itau") || strings.Contains(m, "itaú") || strings.Contains(m, "nubank") || strings.Contains(m, "inter") {
		return "open_finance"
	}

	if strings.Contains(m, "abrir conta") || strings.Contains(m, "criar conta") || strings.Contains(m, "cadastro") ||
		strings.Contains(m, "selfie") || strings.Contains(m, "cpf") || strings.Contains(m, "cnpj") {
		return "criacao_conta"
	}

	return "atendimento_geral"
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MessageRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	agent := routeAgent(req.Message)

	reply := ""

	switch agent {
	case "golpe_med":
		reply = "Olá! Sinto muito pelo ocorrido. Para te ajudar, preciso de algumas informações: qual foi o valor, a data/hora e para qual chave/conta você enviou o Pix?"
	case "open_finance":
		reply = "Entendi! Vamos resolver. Primeiro, copie o link e tente abrir no Chrome/Safari (fora do WhatsApp). Se aparecer erro, me diga qual mensagem aparece."
	case "criacao_conta":
		reply = "Claro! Você está tentando abrir conta ou está com erro no cadastro (ex.: selfie/documento)? Me diga exatamente o que aparece na tela para eu te orientar."
	default:
		reply = "Olá! Eu sou a Aline, assistente virtual do Jota. Para eu te direcionar melhor, você pode me dizer em uma frase qual é a sua necessidade?"
	}

	resp := MessageResponse{
		Reply:  reply,
		Action: "reply",
		Agent:  agent,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/messages", messagesHandler)

	addr := ":8080"
	log.Printf("Server running on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
