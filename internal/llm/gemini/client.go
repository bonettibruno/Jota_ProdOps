package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bonettibruno/Jota_ProdOps/internal/llm"
	"google.golang.org/genai"
)

type Client struct {
	model string
	c     *genai.Client
}

// New initializes a new Gemini client using API keys from environment variables
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

// RouteAgent determines which specialized agent should handle the user request
func (g *Client) RouteAgent(ctx context.Context, traceID string, message string, history []string) (llm.RouterDecision, error) {
	system := `Você é o roteador do Jota. 
Sua saída deve ser EXCLUSIVAMENTE um JSON no formato: {"agent": "nome_do_agente"}

Agentes disponíveis:
- atendimento_geral (assuntos diversos)
- open_finance (conexão de bancos, compartilhamento de dados)
- golpe_med (vítima de golpe, Pix fraudulento, roubo)
- criacao_conta (abertura de conta)`

	var sb strings.Builder
	sb.WriteString("trace_id: " + traceID + "\n\n")
	sb.WriteString("Recent history:\n")
	for i := 0; i < len(history) && i < 6; i++ {
		sb.WriteString("- " + history[i] + "\n")
	}
	sb.WriteString("\nCurrent message:\n" + message + "\n")

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Enforce JSON output via Gemini native configuration
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

	rawText := resp.Text()
	log.Printf("trace=%s event=router_raw_output text=%s", traceID, rawText)

	var dec llm.RouterDecision
	if err := json.Unmarshal([]byte(rawText), &dec); err != nil {
		return llm.RouterDecision{}, fmt.Errorf("router JSON parse failed: %w", err)
	}

	if dec.Agent == "" {
		dec.Agent = "atendimento_geral"
	}

	return dec, nil
}

// GenerateText sends a prompt to the LLM with system instructions and JSON response format
func (g *Client) GenerateText(
	ctx context.Context,
	traceID string,
	systemPrompt string,
	userPrompt string,
) (string, error) {

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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
