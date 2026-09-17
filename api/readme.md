# 💰 Minhas Finanças

Backend for a **personal finance management** application built in **Go** using **microservices**. It allows users to register financial transactions, organize them by category, and create savings goals with progress tracking.

---

## 📌 What the project does

- **Authentication** with JWT and refresh token rotation
- **User registration** with password validation
- **Categories** for income/expenses (e.g., Salary, Groceries, Transport)
- **Financial transactions** with filters by date, type, and category
- **Financial goals** with automatic progress calculation and monthly contribution suggestions
- **Reports** on balance by period and spending by category
- **Asynchronous communication** between services via Kafka

---

## 🧱 Architecture

The project is a **monorepo** with 4 independent microservices, each with its own logical database and communication via HTTP + Kafka.

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

| Service | Port | Responsibility |
|---------|------|----------------|
| **ms_auth** | 4001 | Login, registration, JWT, refresh token |
| **ms_category** | 4002 | CRUD for financial categories |
| **ms_transaction** | 4003 | CRUD for transactions and reports |
| **ms_goal** | 4004 | Financial goals and progress |

---

## 🛠 Stack

**Language & libraries**
- Go 1.27 (with `go.work` for the monorepo)
- chi (HTTP router)
- PostgreSQL + golang-migrate
- Redis (cache)
- Kafka / sarama (messaging)
- JWT (RS256) + bcrypt
- OpenTelemetry (tracing)

**Infrastructure**
- Docker Compose
- PostgreSQL 15, Redis 7, Apache Kafka
- Jaeger (tracing), Prometheus (metrics), Grafana (dashboards)

---

## 🚀 How to run

### Prerequisites

- Go 1.27+
- Docker and Docker Compose
- Make (optional)

### 1. Configure environment variables

```bash
cp .env.example .env
```

### 2. Start the infrastructure (Postgres, Redis, Kafka, Jaeger...)

```bash
docker compose up -d
```

### 3. Run the services

In separate terminals, from the project root:

```bash
cd ms_auth        && go run ./cmd/api
cd ms_category    && go run ./cmd/api
cd ms_goal        && go run ./cmd/api
cd ms_transaction && go run ./cmd/api
```

### 4. Test

Log in to obtain a token:

```bash
curl -X POST http://localhost:4001/v1/auth \
  -H "Content-Type: application/json" \
  -d '{"email":"joao@empresa.com","password":"Senha123A"}'
```

Use the returned `access_token` in subsequent requests:

```bash
curl http://localhost:4004/v1/goals \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

---

## 📡 Main endpoints

### `ms_auth`
- `POST /v1/auth` — login
- `POST /v1/auth/refresh` — refresh tokens
- `POST /v1/auth/logout` — logout
- `POST /v1/users` — registration

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
- `GET /v1/transactions/summary` — balance and spending by category
- `PUT /v1/transactions/{id}`
- `DELETE /v1/transactions/{id}`

### `ms_goal`
- `POST /v1/goals`
- `GET /v1/goals`
- `GET /v1/goals/{id}`
- `GET /v1/goals/report/{id}` — goal report
- `PUT /v1/goals/{id}`
- `DELETE /v1/goals/{id}`

### Infrastructure (all services)
- `GET /health`
- `GET /metrics` (Prometheus)

---

## 📊 Observability

After starting the stack, access:

| Tool | URL | Description |
|------|-----|-------------|
| **Jaeger** | http://localhost:16686 | Distributed tracing |
| **Prometheus** | http://localhost:9090 | Metrics |
| **Grafana** | http://localhost:3000 | Dashboards (admin/admin) |
| **Kafka UI** | http://localhost:8070 | Topic inspection |

---

## 📁 Project structure

```
minhas_financas/
├── shared/             # Shared code (auth, cache, db, httpx, obs...)
├── messaging/          # Generic Kafka Producer/Consumer
├── ms_auth/            # Authentication service
├── ms_category/        # Category service
├── ms_goal/            # Goal service
├── ms_transaction/     # Transaction service
├── compose.yml         # Docker stack
└── go.work             # Go workspace
```

---

## 🔄 Event flow (Kafka)

Services communicate asynchronously:

- **`goal_created`** — ms_goal → ms_category creates a linked category
- **`goal_deleted`** — ms_goal → ms_category removes the category
- **`category_deleted`** — ms_category → ms_transaction removes transactions
- **`transaction_goal_created`** — ms_transaction → ms_goal creates the link
- **`transaction_goal_deleted`** — ms_transaction → ms_goal removes the link

---