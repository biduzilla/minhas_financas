# Minhas Finanças

Fullstack **personal finance management** application. Backend in **Go** with microservices architecture, frontend in **Angular 22** signal-first. Users record transactions (income/expenses), organize them into categories, create goals with automatic progress tracking, and view balance reports.

---

## Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         Browser                              │
│  ┌───────────────────────────────────────────────────────┐  │
│  │            Angular 22 (localhost:4200)                │  │
│  │  Signals · Standalone · Zoneless · Tailwind           │  │
│  └────────────┬──────────────────────────────────────────┘  │
└───────────────┼─────────────────────────────────────────────┘
                │ httpOnly cookies
                │ dev proxy
                ▼
┌─────────────────────────────────────────────────────────────┐
│                  Go Backend (microservices)                  │
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
│    Observability: Jaeger · Prometheus · Grafana             │
└─────────────────────────────────────────────────────────────┘
```

---

## Stack

### Backend
| Layer | Technology |
|---|---|
| Language | Go |
| HTTP | Chi router |
| Database | PostgreSQL |
| Cache | Redis |
| Messaging | Apache Kafka (KRaft mode) |
| Auth | JWT with refresh token rotation |
| Observability | OpenTelemetry + Jaeger + Prometheus + Grafana |
| Containerization | Docker + Docker Compose |

### Frontend
| Layer | Technology |
|---|---|
| Framework | Angular 22 (standalone + signals + zoneless) |
| Styling | Tailwind CSS 4 |
| Forms | Signal Forms |
| HTTP | HttpClient with functional interceptors |
| Testing | Vitest |
| Node | 22 LTS |

---

## Features

- **Authentication** — login, signup, logout, session with httpOnly cookies
- **Dashboard** — financial summary (balance, income, expenses, breakdown by category)
- **Categories** — full CRUD with `input`/`output` type
- **Transactions** — CRUD with filters (type, category, period), sorting, and pagination
- **Goals** — CRUD + progress report with progress bar and monthly suggestion
- **Mirror category** — automatically created via Kafka event when a goal is created

---

## How to run

### Prerequisites

- **Go 1.22+**
- **Node 22 LTS** — `nvm install 22 && nvm use 22`
- **Docker + Docker Compose**
- **Angular CLI** — `npm install -g @angular/cli@latest`

### 1. Infrastructure (Postgres, Redis, Kafka, Jaeger, Prometheus, Grafana)

In the repository root:

```bash
docker compose up -d
```

Starts 7 containers:

| Container | Port | Purpose |
|---|---|---|
| Postgres | 5432 | Database |
| Redis | 6379 | Cache |
| Kafka | 9092, 9094 | Messaging (internal + external) |
| Kafka UI | 8070 | Kafka web interface |
| Jaeger | 16686 | Distributed tracing |
| Prometheus | 9090 | Metrics |
| Grafana | 3001 | Dashboards |

Verify they're up:

```bash
docker compose ps
```

### 2. Go Backend

Each microservice has its own `main.go` and `.env`. Start all four:

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

Or, if you have a Makefile:

```bash
make run-all
```

Verify:

```bash
curl http://localhost:4001/health
curl http://localhost:4002/health
curl http://localhost:4003/health
curl http://localhost:4004/health
```

### 3. Angular Frontend

```bash
cd web
npm install
ng serve
```

Opens at `http://localhost:4200`.

---

## Architecture

### Authentication with httpOnly cookies

Tokens **never** pass through JavaScript. `ms_auth` sets httpOnly cookies on login; the browser attaches them automatically to every subsequent request.

```
[Login]
Browser → POST localhost:4200/api/auth  (proxy)
              ↓
        → POST localhost:4001/v1/auth
        ← 200 { ok: true } + Set-Cookie: access_token=...; HttpOnly
Browser stores the cookie under the "localhost" domain

[Next request]
Browser → GET localhost:4200/api/transactions
         + Cookie: access_token=...   ← automatic
              ↓ (proxy)
        → GET localhost:4003/v1/transactions
         + Cookie: access_token=...   ← proxy forwarded it
        ← 200 [...]
```

**Why it works:** cookies are not scoped by port, only by domain. A `Set-Cookie` from `:4001` is automatically sent to `:4002`, `:4003`, `:4004`. All four microservices read the same cookie with no extra configuration.

**XSS mitigation:** since the token is httpOnly, an injected script (`document.cookie`) can't see it. The browser still sends it normally.

**Refresh token rotation:** each refresh generates a new pair and revokes the previous one. If a revoked refresh is reused (a sign of theft), the **entire family is revoked** — the user must log in again.

### Development proxy

Angular only talks to `/api/*` on `localhost:4200`. `proxy.conf.json` rewrites to each microservice:

```json
{
  "/api/auth":         { "target": "http://localhost:4001", "pathRewrite": { "^/api/auth": "/v1/auth" } },
  "/api/users":        { "target": "http://localhost:4001", "pathRewrite": { "^/api/users": "/v1/users" } },
  "/api/categories":   { "target": "http://localhost:4002", "pathRewrite": { "^/api/categories": "/v1/categories" } },
  "/api/transactions": { "target": "http://localhost:4003", "pathRewrite": { "^/api/transactions": "/v1/transactions" } },
  "/api/goals":        { "target": "http://localhost:4004", "pathRewrite": { "^/api/goals": "/v1/goals" } }
}
```

From the browser's perspective, everything is `localhost:4200` — **same-origin**. No CORS, no preflight, no `SameSite` blocking.

### Inter-service communication

- **Synchronous (REST)** — direct HTTP calls between services (e.g., `ms_transaction` calls `ms_category` to validate a category)
- **Asynchronous (Kafka)** — domain events:
  - `goal_created` → `ms_category` creates mirror category
  - `goal_deleted` → `ms_category` removes mirror category
  - `category_deleted` → `ms_transaction` deletes linked transactions
  - `transaction_goal_created` / `_deleted` → `ms_goal` updates `current_amount`

### Observability

Each service emits OpenTelemetry traces + Prometheus metrics + structured logs (`slog`).

- **Jaeger** — `http://localhost:16686` — distributed traces
- **Prometheus** — `http://localhost:9090` — raw metrics
- **Grafana** — `http://localhost:3001` (admin/admin) — dashboards

A request trace spans all 4 services via the propagated `X-Request-Id`.

---

## Domain

```mermaid
erDiagram
    USER ||--o{ TRANSACTION : owns
    USER ||--o{ CATEGORY : owns
    USER ||--o{ GOAL : owns
    USER ||--o{ REFRESH_TOKEN : has
    CATEGORY ||--o{ TRANSACTION : classifies
    GOAL ||--o| CATEGORY : "creates mirror category"
    GOAL ||--o{ GOAL_TRANSACTION : has
    TRANSACTION ||--o{ GOAL_TRANSACTION : contributes
```

- **USER** — identity and authentication
- **CATEGORY** — classifies transactions as `input` or `output`. Can be linked to a goal
- **TRANSACTION** — financial movement with a required category
- **GOAL** — goal with target amount, deadline, and status (`IN_PROGRESS` / `COMPLETED` / `EXPIRED` / `CANCELED`)
- **GOAL_TRANSACTION** — N:N link between goals and transactions (created via Kafka)
- **REFRESH_TOKEN** — session control with family (rotation)

---

## Main endpoints

### `ms_auth` — :4001

| Method | Path | Description |
|---|---|---|
| POST | `/v1/auth` | Login (sets httpOnly cookies) |
| POST | `/v1/auth/refresh` | Renews tokens (reads refresh cookie) |
| POST | `/v1/auth/logout` | Revokes family + clears cookies |
| GET | `/v1/auth/session` | `{ authenticated: bool }` — always 200 |
| POST | `/v1/users` | Signup |

### `ms_category` — :4002

| Method | Path | Description |
|---|---|---|
| POST | `/v1/categories` | Create |
| GET | `/v1/categories` | List (`page`, `page_size`, `sort`, `type`) |
| GET | `/v1/categories/{id}` | Detail |
| PUT | `/v1/categories/{id}` | Update (includes `version`) |
| DELETE | `/v1/categories/{id}` | Delete (blocked if `goal_id`) |

### `ms_transaction` — :4003

| Method | Path | Description |
|---|---|---|
| POST | `/v1/transactions` | Create |
| GET | `/v1/transactions` | List (`type`, `category_id`, `start_date`, `end_date`) |
| GET | `/v1/transactions/{id}` | Detail |
| GET | `/v1/transactions/summary` | Financial summary |
| PUT | `/v1/transactions/{id}` | Update (includes `version`) |
| DELETE | `/v1/transactions/{id}` | Delete |

### `ms_goal` — :4004

| Method | Path | Description |
|---|---|---|
| POST | `/v1/goals` | Create |
| GET | `/v1/goals` | List (`status` filter in PT-BR) |
| GET | `/v1/goals/{id}` | Detail |
| GET | `/v1/goals/report/{id}` | Progress report |
| PUT | `/v1/goals/{id}` | Update (includes `version`) |
| DELETE | `/v1/goals/{id}` | Delete |

---

## Architectural decisions

### httpOnly cookies instead of `localStorage`

**Why:** `localStorage` is accessible to any JS on the page. An XSS in any npm dependency compromises the session. httpOnly cookies are not readable by JS — mitigates XSS.

**Cost:** requires CSRF protection (`SameSite=Lax` in dev, `Strict` in prod). Since frontend and API are same-origin via proxy/reverse proxy, the cost is zero in practice.

### No BFF

`ms_auth` itself sets the cookies. No intermediate Node/Express layer is needed. A BFF is only required when the backend can't/shouldn't be modified — not the case here.

### No API Gateway

Each microservice has its own port. In dev, the Angular proxy consolidates everything under `localhost:4200`; in production, a reverse proxy (Nginx/Traefik) plays the same role.

### Cache-aware writes with invalidation

The shared `WriteExecutor` invalidates the service cache after every write (INSERT/UPDATE/DELETE) using `context.WithoutCancel` + `context.WithTimeout`, so the invalidation goroutine doesn't die when the HTTP request is canceled.

### Optimistic locking with `version`

Every `PUT` requires `version`. If the backend responds `409 Conflict`, it means someone else modified the resource between the `GET` and the `PUT`. The frontend shows a banner asking the user to reload.

---

## Useful commands

### Backend

```bash
# Build everything
go build ./...

# Run tests
go test ./...

# Run a specific service
cd ms_auth && go run ./cmd/api

# Apply migration (if you have the tool)
make migrate-up
```

### Frontend

```bash
cd web

# Dev server
ng serve

# Production build
ng build

# Unit tests
ng test

# Generate component
ng generate component features/name/name
```

### Infrastructure

```bash
# Start everything
docker compose up -d

# Follow container logs
docker compose logs -f postgres

# Redis flush (dev)
docker exec -it go_redis_minhas_financas redis-cli -a redis_secure_password FLUSHALL

# Postgres CLI
docker exec -it go_postgres_minhas_financas psql -U api_user -d api_db

# List cache keys
docker exec -it go_redis_minhas_financas redis-cli -a redis_secure_password KEYS '*'
```

---

## Environment variables (backend)

Root `.env` (shared) and per-service `.env`:

```env
# Services
AUTH_SERVER_PORT=4001
CATEGORY_SERVER_PORT=4002
TRANSACTION_SERVER_PORT=4003
GOAL_SERVER_PORT=4004

# Database
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

# Observability
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
JAEGER_PORT_UI=16686
```

---

## License

Personal portfolio project. No license defined.
