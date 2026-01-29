package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"

	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
)

type Client struct {
	model string
	c     *genai.Client
}

func New() (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not set")
	}

	c, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}

	model := os.Getenv("GEMINI_MODEL_ROUTER")
	if model == "" {
		model = "gemini-2.0-flash-lite"
	}

	return &Client{model: model, c: c}, nil
}

func (g *Client) RouteAgent(ctx context.Context, traceID string, message string, history []string) (llm.RouterDecision, error) {
	system := `Você é um roteador de atendimento do Jota.
Escolha exatamente UM agente entre:
- atendimento_geral
- criacao_conta
- open_finance
- golpe_med

Responda SOMENTE com JSON válido.`

	var sb strings.Builder
	sb.WriteString("trace_id: " + traceID + "\n\n")
	sb.WriteString("Histórico recente:\n")
	for i := 0; i < len(history) && i < 6; i++ {
		sb.WriteString("- " + history[i] + "\n")
	}
	sb.WriteString("\nMensagem atual:\n" + message + "\n")

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Usando configuração de JSON para garantir saída limpa
	resp, err := g.c.Models.GenerateContent(
		ctx,
		g.model,
		genai.Text(system+"\n\n"+sb.String()),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
		},
	)
	if err != nil {
		return llm.RouterDecision{}, err
	}

	var dec llm.RouterDecision
	if err := json.Unmarshal([]byte(resp.Text()), &dec); err != nil {
		return llm.RouterDecision{}, fmt.Errorf("router JSON parse failed: %w", err)
	}

	return dec, nil
}

// GenerateText atualizado para suportar System Prompt e JSON nativo
func (g *Client) GenerateText(
	ctx context.Context,
	traceID string,
	systemPrompt string,
	userPrompt string,
) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Correção aqui: SystemInstruction espera um *genai.Content
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: systemPrompt},
			},
		},
		ResponseMIMEType: "application/json",
	}

	resp, err := g.c.Models.GenerateContent(
		ctx,
		g.model,
		genai.Text(userPrompt),
		config,
	)
	if err != nil {
		return "", fmt.Errorf("gemini generate text failed: %w", err)
	}

	return resp.Text(), nil
}

// Mantido para compatibilidade caso o RouteAgent ainda precise,
// embora com ResponseMIMEType ele se torne menos necessário.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		if end := strings.LastIndex(s, "```"); end != -1 {
			s = s[:end]
		}
	}
	return strings.TrimSpace(s)
}
