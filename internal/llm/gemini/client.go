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

Responda SOMENTE com JSON válido no formato:
{"agent":"...","confidence":0.0,"reason":"..."}
NÃO use markdown. Responda apenas com o JSON puro, sem cercas e sem texto extra.`

	var sb strings.Builder
	sb.WriteString("trace_id: " + traceID + "\n\n")
	sb.WriteString("Histórico recente:\n")
	for i := 0; i < len(history) && i < 6; i++ {
		sb.WriteString("- " + history[i] + "\n")
	}
	sb.WriteString("\nMensagem atual:\n" + message + "\n")

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	resp, err := g.c.Models.GenerateContent(
		ctx,
		g.model,
		genai.Text(system+"\n\n"+sb.String()),
		nil,
	)
	if err != nil {
		return llm.RouterDecision{}, err
	}

	raw := stripCodeFences(resp.Text())

	var dec llm.RouterDecision
	if err := json.Unmarshal([]byte(raw), &dec); err != nil {
		return llm.RouterDecision{}, fmt.Errorf("router JSON parse failed: %w; raw=%q", err, raw)
	}

	return dec, nil
}

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
