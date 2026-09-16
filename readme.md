# 💰 Minhas Finanças

Backend de uma aplicação de **gestão financeira pessoal** construído em **Go** com **microsserviços**. Permite que usuários cadastrem transações financeiras, organizem-nas por categorias e criem metas de economia com acompanhamento de progresso.

---

## 📌 O que o projeto faz

- **Autenticação** com JWT e refresh token rotation
- **Cadastro de usuários** com validação de senha
- **Categorias** de entrada/saída (ex.: Salário, Mercado, Transporte)
- **Transações** financeiras com filtros por data, tipo e categoria
- **Metas financeiras** com cálculo automático de progresso e sugestão de aporte mensal
- **Relatórios** de saldo por período e gastos por categoria
- **Comunicação assíncrona** entre serviços via Kafka

---

## 🧱 Arquitetura

O projeto é um **monorepo** com 4 microsserviços independentes, cada um com seu próprio banco lógico e comunicação via HTTP + Kafka.

```
┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────────┐
│  ms_auth   │  │ms_category │  │  ms_goal   │  │ ms_transaction │
└─────┬──────┘  └─────┬──────┘  └─────┬──────┘  └───────┬────────┘
      │               │               │                 │
      └───────────────┴───────┬───────┴─────────────────┘
                              │
                    ┌─────────┼─────────┐
                    │         │         │
              ┌─────▼──┐ ┌────▼───┐ ┌───▼────┐
              │Postgres│ │ Redis  │ │ Kafka  │
              └────────┘ └────────┘ └────────┘
```

| Serviço | Porta | Responsabilidade |
|---------|-------|------------------|
| **ms_auth** | 4001 | Login, cadastro, JWT, refresh token |
| **ms_category** | 4002 | CRUD de categorias financeiras |
| **ms_transaction** | 4003 | CRUD de transações e relatórios |
| **ms_goal** | 4004 | Metas financeiras e progresso |

---

## 🛠 Stack

**Linguagem & bibliotecas**
- Go 1.27 (com `go.work` para o monorepo)
- chi (router HTTP)
- PostgreSQL + golang-migrate
- Redis (cache)
- Kafka / sarama (mensageria)
- JWT (RS256) + bcrypt
- OpenTelemetry (tracing)

**Infraestrutura**
- Docker Compose
- PostgreSQL 15, Redis 7, Apache Kafka
- Jaeger (tracing), Prometheus (métricas), Grafana (dashboards)

---

## 🚀 Como rodar

### Pré-requisitos

- Go 1.27+
- Docker e Docker Compose
- Make (opcional)

### 1. Configurar variáveis de ambiente

```bash
cp .env.example .env
```

### 2. Subir a infraestrutura (Postgres, Redis, Kafka, Jaeger...)

```bash
docker compose up -d
```

### 3. Rodar os serviços

Em terminais separados, a partir da raiz do projeto:

```bash
cd ms_auth        && go run ./cmd/api
cd ms_category    && go run ./cmd/api
cd ms_goal        && go run ./cmd/api
cd ms_transaction && go run ./cmd/api
```

### 4. Testar

Faça login para obter um token:

```bash
curl -X POST http://localhost:4001/v1/auth \
  -H "Content-Type: application/json" \
  -d '{"email":"joao@empresa.com","password":"Senha123A"}'
```

Use o `access_token` retornado nas demais requisições:

```bash
curl http://localhost:4004/v1/goals \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

---

## 📡 Principais endpoints

### `ms_auth`
- `POST /v1/auth` — login
- `POST /v1/auth/refresh` — renovar tokens
- `POST /v1/auth/logout` — logout
- `POST /v1/users` — cadastro

### `ms_category`
- `POST /v1/categories`
- `GET /v1/categories`
- `GET /v1/categories/{id}`
- `PUT /v1/categories/{id}`
- `DELETE /v1/categories/{id}`

### `ms_transaction`
- `POST /v1/transactions`
- `GET /v1/transactions`
- `GET /v1/transactions/{id}`
- `GET /v1/transactions/summary` — saldo e gastos por categoria
- `PUT /v1/transactions/{id}`
- `DELETE /v1/transactions/{id}`

### `ms_goal`
- `POST /v1/goals`
- `GET /v1/goals`
- `GET /v1/goals/{id}`
- `GET /v1/goals/report/{id}` — relatório da meta
- `PUT /v1/goals/{id}`
- `DELETE /v1/goals/{id}`

### Infraestrutura (todos os serviços)
- `GET /health`
- `GET /metrics` (Prometheus)

---

## 📊 Observabilidade

Depois de subir a stack, acesse:

| Ferramenta | URL | Descrição |
|------------|-----|-----------|
| **Jaeger** | http://localhost:16686 | Tracing distribuído |
| **Prometheus** | http://localhost:9090 | Métricas |
| **Grafana** | http://localhost:3000 | Dashboards (admin/admin) |
| **Kafka UI** | http://localhost:8070 | Inspeção de tópicos |

---

## 📁 Estrutura resumida

```
minhas_financas/
├── shared/             # Código compartilhado (auth, cache, db, httpx, obs...)
├── messaging/          # Producer/Consumer genéricos do Kafka
├── ms_auth/            # Serviço de autenticação
├── ms_category/        # Serviço de categorias
├── ms_goal/            # Serviço de metas
├── ms_transaction/     # Serviço de transações
├── compose.yml         # Stack Docker
└── go.work             # Workspace Go
```

---

## 🔄 Fluxo de eventos (Kafka)

Os serviços se comunicam de forma assíncrona:

- **`goal_created`** — ms_goal → ms_category cria categoria vinculada
- **`goal_deleted`** — ms_goal → ms_category remove categoria
- **`category_deleted`** — ms_category → ms_transaction remove transações
- **`transaction_goal_created`** — ms_transaction → ms_goal cria vínculo
- **`transaction_goal_deleted`** — ms_transaction → ms_goal remove vínculo

---
