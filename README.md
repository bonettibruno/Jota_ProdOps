# Jota AI — Agentic Orchestration Platform 🚀

O **Jota AI** é uma plataforma de **orquestração de agentes inteligentes** projetada para lidar com fluxos de atendimento complexos, sensíveis e de alta criticidade. Construído em **Go (Golang)**, o sistema prioriza **baixa latência**, **rastreabilidade ponta a ponta** e uma **arquitetura modular**, permitindo a rápida expansão de novas capacidades e especialistas.

---

## 🧠 Arquitetura do Sistema

A plataforma opera sob o conceito de **Agentes Especialistas**.  
Em vez de uma única IA generalista, o sistema utiliza um **orquestrador central** que:

1. Identifica a intenção do usuário  
2. Decide a próxima ação operacional  
3. Executa um **silent handoff** (*transbordo silencioso*) para o agente mais qualificado  

Esse modelo garante respostas mais precisas, previsíveis e alinhadas ao contexto de negócio.

### 🔹 Fluxo Simplificado

```
Canal → Orquestrador → Agente Especialista → Resposta / Próxima Ação
```

---

## ✨ Diferenciais Técnicos

- **RAG (Retrieval‑Augmented Generation)**  
  Recuperação lexical baseada em Markdown que injeta contexto dinâmico **apenas quando necessário**, reduzindo custo e latência.

- **Action‑Driven Engine**  
  O sistema não apenas responde. Ele decide a **próxima ação**:
  - `reply` – responder ao usuário  
  - `ask` – solicitar mais dados  
  - `collect_data` – estruturar informações  
  - `escalate` – acionar intervenção humana  

- **Telemetria de Produção**  
  Métricas nativas para observabilidade completa do comportamento do sistema e dos agentes.

---

## 🛠️ Como Adicionar um Novo Agente

O Jota AI foi desenhado para ser **extensível por design**.  
A criação de um novo agente especialista segue um fluxo simples e padronizado.

### 1️⃣ Criar o *Brain* do Agente

Crie uma nova pasta em:

```
internal/agents/emprestimos/
```

Dentro dela, implemente a interface `core.AgentBrain`, definindo:

- System Prompt do agente  
- Regras de negócio  
- Tipos de ação que ele pode executar  

Exemplo conceitual:
```go
type Brain struct {}

func (b *Brain) Think(ctx core.Context) core.Decision {
    // lógica do agente
}
```

---

### 2️⃣ Registrar o Agente no Orquestrador

No arquivo:

```
internal/api/handlers.go
```

Adicione o novo agente ao mapa de *brains*:

```go
var brains = map[string]core.AgentBrain{
    "emprestimos": &emprestimos.Brain{},
    // outros agentes
}
```

Esse é o único ponto de acoplamento com o orquestrador.

---

### 3️⃣ Atualizar a Base de Conhecimento (RAG)

Edite o arquivo:

```
kb/RAG_JOTA_RESUMIDO.md
```

Adicione uma nova seção com cabeçalho Markdown:

```md
# Empréstimos
Conteúdo relevante para o agente...
```

O motor de RAG irá **indexar automaticamente** esse conteúdo e disponibilizá‑lo apenas para o agente quando necessário.

---

## 🚀 Operação e Monitoramento

A plataforma foi construída com foco em **observabilidade real de produção**.

### 🔍 Rastreabilidade

- Cada requisição recebe um `X-Trace-Id`
- O identificador acompanha toda a execução, mesmo em transbordos entre agentes

### ❤️ Health Check

```
GET /health
```

Utilizado para monitoramento por clusters, load balancers e orquestradores.

### 📊 Métricas

```
GET /metrics
```

Principais indicadores:

- Volume total de requisições  
- Taxa de transbordo entre agentes  
- Tempo médio de resposta por agente  
- Número de `escalates` (intervenção humana)

---

## 📦 Deploy

O projeto é **100% dockerizado**, utilizando **multi‑stage builds** para gerar imagens leves, seguras e prontas para produção.

### ▶️ Subir a Plataforma

```bash
docker compose up --build
```

O servidor será exposto em:

```
http://localhost:8080
```

Pronto para integração via **Webhooks** com canais como:

- WhatsApp  
- Webchat  
- Aplicações Mobile  

---

## 🛠️ Tecnologias Utilizadas

- **Linguagem:** Go (Golang) 1.24  
- **LLM Client:** Google Gemini API  
- **Infraestrutura:** Docker / Docker Compose  
- **Contexto:** RAG baseado em Markdown  
- **Observabilidade:** Logs estruturados + métricas nativas  

---

## 📌 Visão Geral

O Jota AI não é apenas um chatbot.  
É uma **plataforma de decisão agentica**, pensada para ambientes onde **controle, previsibilidade e rastreabilidade** são tão importantes quanto inteligência.
