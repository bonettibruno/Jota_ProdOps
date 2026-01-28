package core

import (
	"fmt"
	"strings"

	"github.com/bonettibruno/Jota_ProdOps/internal/rag"
)

type Citation struct {
	Source  string `json:"source"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

func BuildReplyWithRAG(agent string, userMessage string, retriever *rag.Retriever) (string, []Citation) {
	// Busca top 2 trechos
	chunks := retriever.Search(userMessage, 2)

	citations := make([]Citation, 0, len(chunks))
	var contextLines []string

	for _, c := range chunks {
		snippet := firstNChars(strings.TrimSpace(c.Content), 240)
		citations = append(citations, Citation{
			Source:  "RAG_JOTA_RESUMIDO.md",
			Title:   c.Title,
			Snippet: snippet,
		})
		contextLines = append(contextLines, fmt.Sprintf("- %s: %s", c.Title, snippet))
	}

	// Resposta simples: mantém o “tom” por agente e injeta contexto do doc
	base := AgentReply(agent)

	if len(contextLines) == 0 {
		// sem achados: devolve base + pedir clarificação
		return base + " Se puder, me diga mais detalhes para eu te orientar da forma certa.", nil
	}

	reply := base + "\n\nBaseado na nossa base de conhecimento:\n" + strings.Join(contextLines, "\n")
	return reply, citations
}

func firstNChars(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
