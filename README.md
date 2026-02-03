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

A plataforma expõe um endpoint nativo de métricas em `GET /metrics`. Este endpoint fornece dados brutos em tempo real, permitindo a extração dos seguintes KPIs operacionais:

- **Volumetria Total:** (`total_requests`) Quantidade total de interações processadas.
- **Eficiência de Triagem:** (`total_handoffs`) Volume de trocas de contexto entre agentes especialistas.
- **Taxa de Escalada Humana:** (`total_escalates`) Identificação de casos críticos que exigiram intervenção manual.
- **Distribuição de Carga:** (`requests_by_agent`) Monitoramento de qual especialista está sendo mais demandado (ex: Golpe MED vs. Atendimento Geral).

> **Nota de ProdOps:** Os logs da aplicação também registram a latência individual de cada requisição (`latency=Xms`), permitindo a análise de performance e gargalos de processamento por agente.

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

## 🧪 Como Testar os Fluxos Principais

Após subir o container (`docker compose up`), é possível validar a inteligência dos agentes, o roteamento do orquestrador e a execução das **Actions** utilizando chamadas `curl`.

---

## 🧪 Caso de Teste Exploratório — Conversa Livre com Transbordo Automático

Este caso de teste foi pensado para **exploração manual** do comportamento agentico do Jota AI.  
A ideia é conversar naturalmente com a IA e observar como o **orquestrador realiza o handoff silencioso** entre agentes, sem que o usuário precise saber qual especialista está ativo.

Durante a conversa, o sistema pode:
- Iniciar em um agente generalista
- Transbordar automaticamente para um agente de segurança
- Evoluir para um agente operacional ou de suporte
- Decidir sozinho quando escalar ou coletar mais dados

### Objetivo do teste
- Validar **mudança automática de agentes**
- Avaliar coerência do contexto entre mensagens
- Observar decisões de Action (`reply`, `ask`, `call_api`, `escalate`)
- Conferir telemetria de `handoffs`

---

### Exemplo de Conversa (Passo a Passo)

#### 1️⃣ Início — Conversa aberta
```bash
curl -X POST http://localhost:8080/messages -H "Content-Type: application/json" -d '{
  "conversation_id": "user-explore-001",
  "message": "Oi, acho que aconteceu algo estranho com uma transferência que fiz hoje."
}'
```

**Comportamento esperado:**
- IA responde com perguntas de esclarecimento
- `action`: `"ask"`
- Agente inicial (generalista ou triagem)

---

#### 2️⃣ Indício de fraude
```bash
curl -X POST http://localhost:8080/messages -H "Content-Type: application/json" -d '{
  "conversation_id": "user-explore-001",
  "message": "Foi um Pix e a pessoa sumiu depois que recebeu."
}'
```

**Comportamento esperado:**
- Handoff silencioso para **Agente de Segurança**
- Contexto preservado
- `action`: `"ask"` ou `"reply"`

---

#### 3️⃣ Confirmação de golpe
```bash
curl -X POST http://localhost:8080/messages -H "Content-Type: application/json" -d '{
  "conversation_id": "user-explore-001",
  "message": "Sim, eu já registrei um boletim de ocorrência."
}'
```

**Comportamento esperado:**
- IA reconhece pré-requisitos do MED
- `action`: `"call_api"`
- Execução do fluxo de protocolo MED
- Incremento de `total_handoffs` e `requests_by_agent=seguranca`

---

#### 4️⃣ Continuidade da conversa
```bash
curl -X POST http://localhost:8080/messages -H "Content-Type: application/json" -d '{
  "conversation_id": "user-explore-001",
  "message": "Tem mais alguma coisa que eu precise fazer agora?"
}'
```

**Comportamento esperado:**
- IA mantém o agente correto ativo
- Resposta orientativa clara
- Possível `action`: `"reply"` ou `"collect_data"`

---


## 🛠️ Tecnologias Utilizadas

- **Linguagem:** Go (Golang) 1.24  
- **LLM Client:** Google Gemini API  
- **Infraestrutura:** Docker / Docker Compose  
- **Contexto:** RAG baseado em Markdown  
- **Observabilidade:** Logs estruturados + métricas nativas  

---


## 📜 Script de Teste de carga para usuário enviando várias mensagens antes de receber uma resposta

```bash
#!/bin/bash

CONV_ID="USER_ANSIOSO_123"

curl -X POST http://localhost:8080/messages \
  -d "{\"conversation_id\": \"$CONV_ID\", \"message\": \"Oi\"}" &

curl -X POST http://localhost:8080/messages \
  -d "{\"conversation_id\": \"$CONV_ID\", \"message\": \"Cai num golpe aqui\"}" &

curl -X POST http://localhost:8080/messages \
  -d "{\"conversation_id\": \"$CONV_ID\", \"message\": \"Foi um pix de 200 reais\"}" &

wait

echo -e "\nTeste finalizado. Verifique o [MEMORY DUMP] no terminal do servidor."
```

---

## 📝 Próximos Passos (To‑Do)

- **Resolver múltiplas mensagens**  
  Implementar um mecanismo de espera (ex: aguardar 1–2 segundos após a última mensagem antes de responder).

- **Persistência Estruturada**  
  Migrar da memória volátil para banco de dados.

- **Extração de Dados e APIs**  
  Capturar automaticamente dados do chat (valor, chave Pix) e disparar chamadas reais (ex: API do Formulário MED).

- **Integrações Externas**  
  Conectar o gatilho de `escalate` ao Zendesk para o atendimento humano.

- **Integração com WhatsApp**  
  Configurar Webhooks para mensagens reais e respostas via API oficial.

- **Observabilidade para Humanos**  
  Criar painel ou logs estruturados permitindo que o atendente humano visualize todo o histórico gerado pela IA antes de assumir o caso.