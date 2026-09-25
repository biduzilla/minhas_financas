# Minhas Finanças

Aplicação fullstack de **gestão financeira pessoal**. Backend em **Go** com arquitetura de microsserviços, frontend em **Angular 22** signal-first. O usuário registra transações (entradas/saídas), organiza em categorias, cria metas com progresso automático e consulta relatórios de balanço.

---

## Visão geral

```
┌─────────────────────────────────────────────────────────────┐
│                         Browser                              │
│  ┌───────────────────────────────────────────────────────┐  │
│  │            Angular 22 (localhost:4200)                │  │
│  │  Signals · Standalone · Zoneless · Tailwind           │  │
│  └────────────┬──────────────────────────────────────────┘  │
└───────────────┼─────────────────────────────────────────────┘
                │ cookies httpOnly
                │ proxy de dev
                ▼
┌─────────────────────────────────────────────────────────────┐
│                  Backend Go (microsserviços)                 │
│                                                              │
│  ┌──────────┐  ┌────────────┐  ┌──────────────┐  ┌────────┐ │
│  │ ms_auth  │  │ms_category │  │ms_transaction│  │ms_goal │ │
│  │  :4001   │  │   :4002    │  │    :4003     │  │ :4004  │ │
│  └────┬─────┘  └─────┬──────┘  └──────┬───────┘  └───┬────┘ │
│       │              │                │              │      │
│       └──────────────┴────────┬───────┴──────────────┘      │
│                               │                              │
│         ┌─────────────────────┼─────────────────────┐       │
│         ▼                     ▼                     ▼       │
│    ┌─────────┐          ┌──────────┐          ┌─────────┐  │
│    │Postgres │          │  Redis   │          │  Kafka  │  │
│    └─────────┘          └──────────┘          └─────────┘  │
│                                                              │
│    Observabilidade: Jaeger · Prometheus · Grafana           │
└─────────────────────────────────────────────────────────────┘
```

---

## Stack

### Backend
| Camada | Tecnologia |
|---|---|
| Linguagem | Go |
| HTTP | Chi router |
| Banco | PostgreSQL |
| Cache | Redis |
| Mensageria | Apache Kafka (KRaft mode) |
| Auth | JWT com refresh token rotation |
| Observabilidade | OpenTelemetry + Jaeger + Prometheus + Grafana |
| Containerização | Docker + Docker Compose |

### Frontend
| Camada | Tecnologia |
|---|---|
| Framework | Angular 22 (standalone + signals + zoneless) |
| Estilos | Tailwind CSS 4 |
| Formulários | Signal Forms |
| HTTP | HttpClient com interceptors funcionais |
| Testes | Vitest |
| Node | 22 LTS |

---

## Funcionalidades

- **Autenticação** — login, signup, logout, sessão com cookies httpOnly
- **Dashboard** — resumo financeiro (saldo, entradas, saídas, breakdown por categoria)
- **Categorias** — CRUD completo com tipo `input`/`output`
- **Transações** — CRUD com filtros (tipo, categoria, período), ordenação e paginação
- **Metas** — CRUD + relatório com barra de progresso e sugestão mensal
- **Categoria espelho** — criada automaticamente via evento Kafka ao criar uma meta

---

## Como rodar

### Pré-requisitos

- **Go 1.22+**
- **Node 22 LTS** — `nvm install 22 && nvm use 22`
- **Docker + Docker Compose**
- **Angular CLI** — `npm install -g @angular/cli@latest`

### 1. Infraestrutura (Postgres, Redis, Kafka, Jaeger, Prometheus, Grafana)

No diretório raiz:

```bash
docker compose up -d
```

Sobe 7 containers:

| Container | Porta | Propósito |
|---|---|---|
| Postgres | 5432 | Banco de dados |
| Redis | 6379 | Cache |
| Kafka | 9092, 9094 | Mensageria (interna + externa) |
| Kafka UI | 8070 | Interface web pro Kafka |
| Jaeger | 16686 | Tracing distribuído |
| Prometheus | 9090 | Métricas |
| Grafana | 3001 | Dashboards |

Confere que subiu:

```bash
docker compose ps
```

### 2. Backend Go

Cada microserviço tem seu próprio `main.go` e `.env`. Sobe os quatro:

```bash
# Terminal 1
cd ms_auth && go run ./cmd/api

# Terminal 2
cd ms_category && go run ./cmd/api

# Terminal 3
cd ms_transaction && go run ./cmd/api

# Terminal 4
cd ms_goal && go run ./cmd/api
```

Ou, se você tiver um Makefile:

```bash
make run-all
```

Confere:

```bash
curl http://localhost:4001/health
curl http://localhost:4002/health
curl http://localhost:4003/health
curl http://localhost:4004/health
```

### 3. Frontend Angular

```bash
cd web
npm install
ng serve
```

Abre em `http://localhost:4200`.

---

## Arquitetura

### Autenticação com cookies httpOnly

Os tokens **nunca** passam pelo JavaScript. O `ms_auth` seta cookies httpOnly no login; o browser anexa automaticamente em toda requisição subsequente.

```
[Login]
Browser → POST localhost:4200/api/auth  (proxy)
              ↓
        → POST localhost:4001/v1/auth
        ← 200 { ok: true } + Set-Cookie: access_token=...; HttpOnly
Browser guarda o cookie no domínio "localhost"

[Próxima request]
Browser → GET localhost:4200/api/transactions
         + Cookie: access_token=...   ← automático
              ↓ (proxy)
        → GET localhost:4003/v1/transactions
         + Cookie: access_token=...   ← proxy repassou
        ← 200 [...]
```

**Por que funciona:** cookie não é escopado por porta, só por domínio. Um `Set-Cookie` do `:4001` é enviado automaticamente pro `:4002`, `:4003`, `:4004`. Os quatro microserviços leem o mesmo cookie sem configuração extra.

**Mitigação de XSS:** como o token é httpOnly, um script injetado (`document.cookie`) não o vê. O browser continua enviando normalmente.

**Refresh token rotation:** cada refresh gera um novo par e revoga o anterior. Se um refresh já revogado for reutilizado (indício de roubo), **toda a família é revogada** — o usuário precisa logar de novo.

### Proxy de desenvolvimento

O Angular fala apenas com `/api/*` em `localhost:4200`. O `proxy.conf.json` reescreve pra cada microserviço:

```json
{
  "/api/auth":         { "target": "http://localhost:4001", "pathRewrite": { "^/api/auth": "/v1/auth" } },
  "/api/users":        { "target": "http://localhost:4001", "pathRewrite": { "^/api/users": "/v1/users" } },
  "/api/categories":   { "target": "http://localhost:4002", "pathRewrite": { "^/api/categories": "/v1/categories" } },
  "/api/transactions": { "target": "http://localhost:4003", "pathRewrite": { "^/api/transactions": "/v1/transactions" } },
  "/api/goals":        { "target": "http://localhost:4004", "pathRewrite": { "^/api/goals": "/v1/goals" } }
}
```

Pro browser, tudo é `localhost:4200` — **same-origin**. Sem CORS, sem preflight, sem `SameSite` bloqueando.

### Comunicação entre serviços

- **Síncrona (REST)** — chamadas diretas entre serviços via HTTP (ex: `ms_transaction` chama `ms_category` pra validar categoria)
- **Assíncrona (Kafka)** — eventos de domínio:
  - `goal_created` → `ms_category` cria categoria espelho
  - `goal_deleted` → `ms_category` remove categoria espelho
  - `category_deleted` → `ms_transaction` apaga transações vinculadas
  - `transaction_goal_created` / `_deleted` → `ms_goal` atualiza `current_amount`

### Observabilidade

Cada serviço emite traces OpenTelemetry + métricas Prometheus + logs estruturados (`slog`).

- **Jaeger** — `http://localhost:16686` — traces distribuídos
- **Prometheus** — `http://localhost:9090` — métricas brutas
- **Grafana** — `http://localhost:3001` (admin/admin) — dashboards

Trace de uma requisição atravessa os 4 serviços por `X-Request-Id` propagado.

---

## Domínio

```mermaid
erDiagram
    USER ||--o{ TRANSACTION : owns
    USER ||--o{ CATEGORY : owns
    USER ||--o{ GOAL : owns
    USER ||--o{ REFRESH_TOKEN : has
    CATEGORY ||--o{ TRANSACTION : classifies
    GOAL ||--o| CATEGORY : "cria categoria espelho"
    GOAL ||--o{ GOAL_TRANSACTION : has
    TRANSACTION ||--o{ GOAL_TRANSACTION : contributes
```

- **USER** — identidade e autenticação
- **CATEGORY** — classifica transações como `input` ou `output`. Pode estar vinculada a uma meta
- **TRANSACTION** — movimentação financeira com categoria obrigatória
- **GOAL** — meta com valor-alvo, deadline e status (`IN_PROGRESS` / `COMPLETED` / `EXPIRED` / `CANCELED`)
- **GOAL_TRANSACTION** — vínculo N:N entre metas e transações (criado via Kafka)
- **REFRESH_TOKEN** — controle de sessão com família (rotação)

---

## Endpoints principais

### `ms_auth` — :4001

| Método | Path | Descrição |
|---|---|---|
| POST | `/v1/auth` | Login (seta cookies httpOnly) |
| POST | `/v1/auth/refresh` | Renova tokens (lê cookie de refresh) |
| POST | `/v1/auth/logout` | Revoga família + limpa cookies |
| GET | `/v1/auth/session` | `{ authenticated: bool }` — sempre 200 |
| POST | `/v1/users` | Signup |

### `ms_category` — :4002

| Método | Path | Descrição |
|---|---|---|
| POST | `/v1/categories` | Criar |
| GET | `/v1/categories` | Listar (`page`, `page_size`, `sort`, `type`) |
| GET | `/v1/categories/{id}` | Detalhe |
| PUT | `/v1/categories/{id}` | Atualizar (inclui `version`) |
| DELETE | `/v1/categories/{id}` | Remover (bloqueado se `goal_id`) |

### `ms_transaction` — :4003

| Método | Path | Descrição |
|---|---|---|
| POST | `/v1/transactions` | Criar |
| GET | `/v1/transactions` | Listar (`type`, `category_id`, `start_date`, `end_date`) |
| GET | `/v1/transactions/{id}` | Detalhe |
| GET | `/v1/transactions/summary` | Resumo financeiro |
| PUT | `/v1/transactions/{id}` | Atualizar (inclui `version`) |
| DELETE | `/v1/transactions/{id}` | Deletar |

### `ms_goal` — :4004

| Método | Path | Descrição |
|---|---|---|
| POST | `/v1/goals` | Criar |
| GET | `/v1/goals` | Listar (`status` em PT-BR) |
| GET | `/v1/goals/{id}` | Detalhe |
| GET | `/v1/goals/report/{id}` | Relatório de progresso |
| PUT | `/v1/goals/{id}` | Atualizar (inclui `version`) |
| DELETE | `/v1/goals/{id}` | Deletar |

---

## Decisões arquiteturais

### Cookies httpOnly em vez de `localStorage`

**Por que:** `localStorage` é acessível por qualquer JS na página. Um XSS em qualquer dependência npm derruba a sessão. Cookie httpOnly não é lido por JS — mitiga XSS.

**Custo:** precisa de proteção CSRF (`SameSite=Lax` em dev, `Strict` em prod). Como o front e a API são same-origin via proxy/reverse proxy, o custo é zero na prática.

### Sem BFF

O próprio `ms_auth` seta os cookies. Não precisa de camada intermediária Node/Express. Um BFF só é necessário quando o backend não pode/pode ser alterado — não é o caso.

### Sem API Gateway

Cada microserviço tem sua própria porta. Em dev, o proxy do Angular consolida em `localhost:4200`; em produção, um reverse proxy (Nginx/Traefik) faz o mesmo papel.

### Cache-aware writes com invalidação

O `WriteExecutor` do shared invalida o cache do serviço após cada write (INSERT/UPDATE/DELETE) usando `context.WithoutCancel` + `context.WithTimeout`, pra que a goroutine de invalidação não morra quando a request HTTP for cancelada.

### Optimistic locking com `version`

Todo `PUT` exige `version`. Se o backend responder `409 Conflict`, significa que alguém alterou o recurso entre o `GET` e o `PUT`. O front mostra um banner pedindo recarregar.

---

## Comandos úteis

### Backend

```bash
# Build de tudo
go build ./...

# Rodar testes
go test ./...

# Rodar um serviço específico
cd ms_auth && go run ./cmd/api

# Aplicar migration (se tiver ferramenta)
make migrate-up
```

### Frontend

```bash
cd web

# Dev server
ng serve

# Build produção
ng build

# Testes unitários
ng test

# Gerar componente
ng generate component features/nome/nome
```

### Infraestrutura

```bash
# Sobe tudo
docker compose up -d

# Ver logs de um container
docker compose logs -f postgres

# Flush do Redis (dev)
docker exec -it go_redis_minhas_financas redis-cli -a redis_secure_password FLUSHALL

# Postgres CLI
docker exec -it go_postgres_minhas_financas psql -U api_user -d api_db

# Ver chaves do cache
docker exec -it go_redis_minhas_financas redis-cli -a redis_secure_password KEYS '*'
```

---

## Variáveis de ambiente (backend)

Arquivo `.env` na raiz (compartilhado) e `.env` por serviço:

```env
# Serviços
AUTH_SERVER_PORT=4001
CATEGORY_SERVER_PORT=4002
TRANSACTION_SERVER_PORT=4003
GOAL_SERVER_PORT=4004

# Banco
POSTGRES_USER=api_user
POSTGRES_PASSWORD=api_password
POSTGRES_DB=api_db
DB_DSN=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=redis_secure_password

# Kafka
KAFKA_PORT_INTERNAL=9092
KAFKA_PORT_EXTERNAL=9094

# Observabilidade
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
JAEGER_PORT_UI=16686
```

---

## Licença

Projeto pessoal de portfólio. Sem licença definida.