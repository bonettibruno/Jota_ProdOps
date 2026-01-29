# 🚀 Jota ProdOps – Multi‑Agent System

Este projeto implementa um **sistema de atendimento inteligente baseado em múltiplos agentes de IA**, desenvolvido para o desafio **Jota ProdOps**. A arquitetura foi pensada para escalar especialistas, manter o código organizado e garantir que cada cliente seja atendido sempre pelo **agente mais adequado**, sem fricção ou respostas duplicadas.

---

## 🧠 Visão Geral da Arquitetura

O sistema segue princípios sólidos de engenharia de software, priorizando **baixa complexidade operacional**, **clareza de responsabilidades** e **facilidade de evolução**.

### 🔹 Conceitos‑chave

- **Multi‑Agent System**: cada agente representa um especialista (ex.: atendimento geral, criação de conta, investimentos).
- **Handoff Silencioso**: a troca de agente ocorre internamente, na mesma requisição HTTP, sem que o cliente perceba.
- **RAG (Retrieval‑Augmented Generation)**: contexto extraído de arquivos Markdown oficiais para respostas mais precisas.
- **Baixo Acoplamento & Alta Coesão**: cada agente tem seu próprio “cérebro” isolado.

---

## 🏗️ Decisões Técnicas

### ✅ Modularidade
Cada agente vive em sua própria pasta dentro de `internal/agents`, contendo apenas o código necessário para sua função.

### 🔁 Handoff Recursivo
O **orquestrador** (`handlers.go`) é responsável por:
- Identificar o agente correto
- Transferir internamente o controle
- Executar o agente especialista **na mesma requisição**

Isso evita múltiplas respostas e simplifica o fluxo.

### 🔌 Inversão de Dependência
O pacote `core` define **interfaces** compartilhadas, evitando dependências circulares entre:
- API
- Agentes
- LLM
- RAG

### 📚 RAG em Markdown
O sistema utiliza arquivos `.md` como base de conhecimento, permitindo:
- Versionamento simples
- Fácil auditoria
- Atualização sem recompilar lógica de IA

---

## 📁 Estrutura de Pastas

```text
cmd/
 └── server/
     └── main.go

internal/
 ├── api/
 │   └── handlers.go        # Orquestrador e handoff
 ├── agents/
 │   ├── atendimento/
 │   ├── criacaoconta/
 │   └── investimentos/
 ├── core/
 │   ├── interfaces.go      # Contratos (AgentBrain, ActionPlan, etc.)
 │   └── conversation.go
 ├── llm/
 │   └── client.go
 └── rag/
     └── retriever.go
```

---

## 🛠️ Como Adicionar um Novo Agente

### 1️⃣ Criar a pasta do agente

```bash
mkdir internal/agents/investimentos
```

---

### 2️⃣ Implementar o Brain

Crie o arquivo `brain.go` dentro da pasta do agente:

```go
package investimentos

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/bonettibruno/Jota_ProdOps/internal/core"
    "github.com/bonettibruno/Jota_ProdOps/internal/llm"
)

type Brain struct{}

func (b *Brain) Run(
    ctx context.Context,
    client any,
    traceID string,
    history []core.ChatMessage,
    userMessage string,
    ragContext string,
) (core.ActionPlan, error) {

    // Conversão do client genérico para o cliente LLM
    llmClient := client.(llm.Client)

    system := fmt.Sprintf(
        "Você é o especialista em Investimentos. Contexto oficial: %s",
        ragContext,
    )

    raw, err := llmClient.GenerateText(ctx, traceID, system, userMessage)
    if err != nil {
        return core.ActionPlan{}, err
    }

    var plan core.ActionPlan
    if err := json.Unmarshal([]byte(raw), &plan); err != nil {
        return core.ActionPlan{}, fmt.Errorf("erro no unmarshal: %w", err)
    }

    return plan, nil
}
```

---

### 3️⃣ Registrar o agente no Orquestrador

No arquivo `internal/api/handlers.go`, adicione o novo agente ao mapa de cérebros:

```go
var brains = map[string]core.AgentBrain{
    "atendimento_geral": &atendimento.Brain{},
    "criacao_conta":     &criacaoconta.Brain{},
    "investimentos":     &investimentos.Brain{},
}
```

---

### 4️⃣ Instruir a Transferência de Agente

No `brain.go` do **atendimento geral**, inclua o novo agente como opção válida no campo `change_agent` do prompt.

Isso permite que a IA saiba que pode transferir a conversa para o especialista correto.

---

## 🚀 Como Executar o Projeto

### 🔑 Configurar a API Key

Coloque suas credenciais do Google AI Studio num arquivo .env, seguindo o exemplo. Além disso, escolha o modelo de IA e a porta a ser utilizada.

---

### ▶️ Rodar o servidor

```bash
go run cmd/server/main.go
```

---

### 🧪 Testar a API

```bash
curl -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": "u1",
    "message": "quero investir"
  }'
```

---

## 🎯 Benefícios da Arquitetura

- Escalável para novos agentes
- Fácil manutenção
- Separação clara de responsabilidades
- Ideal para ambientes produtivos e regulados
- Excelente base para evoluir para **ProdOps**, **FinOps** ou **Open Finance**

---

📌 **Projeto desenvolvido para o desafio Jota ProdOps**
