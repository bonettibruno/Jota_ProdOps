# ESTÁGIO 1: Build
FROM golang:1.24-alpine AS builder

# Define o diretório de trabalho
WORKDIR /app

# Copia os arquivos de dependências primeiro (otimiza o cache do Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Compila o binário estático (CGO_ENABLED=0 garante que rode em qualquer linux)
RUN CGO_ENABLED=0 GOOS=linux go build -o jota-server ./cmd/server/main.go

# ESTÁGIO 2: Execução (Imagem final limpa)
FROM alpine:latest

WORKDIR /app

# Instala certificados CA (necessário para chamadas HTTPS da API do Gemini)
RUN apk --no-cache add ca-certificates

# Copia apenas o binário do estágio de build
COPY --from=builder /app/jota-server .

# Copia a base de conhecimento (RAG)
COPY --from=builder /app/kb ./kb

# Expõe a porta do servidor
EXPOSE 8080

# Comando para rodar a aplicação
CMD ["./jota-server"]