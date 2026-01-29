package rag

import (
	"os"
	"regexp"
	"sort"
	"strings"
)

type Chunk struct {
	Title   string
	Content string
	Score   int
}

type Retriever struct {
	path     string
	fullText string
	chunks   []Chunk
}

func NewRetriever(path string) (*Retriever, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	text := string(b)
	chunks := splitMarkdownByHeadings(text)

	return &Retriever{
		path:     path,
		fullText: text,
		chunks:   chunks,
	}, nil
}

func (r *Retriever) Search(query string, topK int) []Chunk {
	qTokens := tokenize(query)
	if len(qTokens) == 0 {
		return nil
	}

	results := make([]Chunk, 0, len(r.chunks))
	for _, c := range r.chunks {
		score := scoreChunk(qTokens, c.Title+" "+c.Content)
		if score > 0 {
			cc := c
			cc.Score = score
			results = append(results, cc)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topK <= 0 || topK > len(results) {
		topK = len(results)
	}
	return results[:topK]
}

func splitMarkdownByHeadings(md string) []Chunk {
	lines := strings.Split(md, "\n")
	var chunks []Chunk

	currentTitle := "Documento"
	var buf []string

	headingRe := regexp.MustCompile(`^\s{0,3}(#{1,6})\s+(.+?)\s*$`)

	flush := func() {
		content := strings.TrimSpace(strings.Join(buf, "\n"))
		if content != "" {
			chunks = append(chunks, Chunk{
				Title:   currentTitle,
				Content: content,
			})
		}
		buf = nil
	}

	for _, line := range lines {
		if m := headingRe.FindStringSubmatch(line); m != nil {
			flush()
			currentTitle = strings.TrimSpace(m[2])
			continue
		}
		buf = append(buf, line)
	}
	flush()

	return chunks
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	// remove pontuação básica
	re := regexp.MustCompile(`[^\p{L}\p{N}\s]+`)
	s = re.ReplaceAllString(s, " ")
	parts := strings.Fields(s)

	// remove stopwords básicas (pode melhorar depois)
	stop := map[string]bool{
		"o": true, "a": true, "os": true, "as": true, "de": true, "da": true, "do": true,
		"e": true, "em": true, "para": true, "por": true, "um": true, "uma": true,
		"nao": true, "não": true, "com": true, "no": true, "na": true, "que": true,
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) < 2 || stop[p] {
			continue
		}
		out = append(out, p)
	}
	return out
}

func scoreChunk(qTokens []string, text string) int {
	t := strings.ToLower(text)
	score := 0
	for _, tok := range qTokens {
		if strings.Contains(t, tok) {
			score++
		}
	}
	return score
}

func (r *Retriever) AsText() string {
	if r == nil {
		return ""
	}
	return r.fullText
}

func (r *Retriever) SearchAsText(query string, topK int) string {
	chunks := r.Search(query, topK)
	if len(chunks) == 0 {
		return "Nenhuma informação relevante encontrada na base de conhecimento."
	}

	var sb strings.Builder
	sb.WriteString("Contexto extraído da Base de Conhecimento:\n")
	for _, c := range chunks {
		sb.WriteString("\n--- " + c.Title + " ---\n")
		sb.WriteString(c.Content + "\n")
	}
	return sb.String()
}
