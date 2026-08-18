# ==============================================
# Makefile para microsserviços Go + Docker
# ==============================================

# Serviço padrão quando não especificado
SERVICE ?= ms_auth

# Lista de serviços para alvos de build/teste futuros
SERVICES := ms_auth

# Infraestrutura que sobe junto com o serviço local
# (adicione ou remova conforme necessário)
INFRA := postgres redis kafka jaeger prometheus grafana

# Pacote de teste padrão
TEST_PKG ?= ./internal/features

# Variáveis de ambiente para o OpenTelemetry
OTEL_ENV := OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 OTEL_EXPORTER_OTLP_INSECURE=true

# ----------------------------------------------
# Alvos principais
# ----------------------------------------------

.PHONY: help run dev infra-up infra-down down logs ps test test-verbose test-cover test-service build

help: ## Mostra esta ajuda
	@echo "Comandos disponíveis:"
	@echo "  make run              - Sobe infra e roda o serviço $(SERVICE) localmente"
	@echo "  make run-only         - Roda o serviço $(SERVICE) localmente"
	@echo "  make dev              - Sobe infra e roda o serviço (modo desenvolvimento)"
	@echo "  make infra-up         - Sobe apenas os containers de infraestrutura"
	@echo "  make infra-down       - Derruba apenas os containers de infraestrutura"
	@echo "  make down             - Derruba todos os containers"
	@echo "  make logs             - Acompanha logs dos containers"
	@echo "  make ps               - Mostra status dos containers"
	@echo "  make test             - Roda todos os testes do ms_auth"
	@echo "  make test-verbose     - Testes com saída detalhada"
	@echo "  make test-cover       - Testes com cobertura"
	@echo "  make test-service     - Testa um serviço específico (SERVICE=ms_auth)"
	@echo "  make build            - Compila o serviço $(SERVICE)"

# Sobe a infraestrutura e depois roda o serviço localmente
run: infra-up
	@echo "==> Iniciando servico $(SERVICE)..."
	cd $(SERVICE) && go run ./cmd/api

# Executa somente o serviço local (sem subir infra)
run-only:
	@echo "==> Executando $(SERVICE) localmente..."
	cd $(SERVICE) && OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 OTEL_EXPORTER_OTLP_INSECURE=true go run ./cmd/api

# Alias para run (mesma coisa, nome semântico)
dev: run

# Sobe apenas os containers de infraestrutura (sem a aplicação)
infra-up:
	@echo "==> Subindo infraestrutura (${INFRA})..."
	docker compose up -d $(INFRA)

# Derruba apenas a infraestrutura
infra-down:
	@echo "==> Derrubando infraestrutura..."
	docker compose stop $(INFRA)

# Derruba todos os containers do projeto
down:
	@echo "==> Derrubando todos os containers..."
	docker compose down

# Logs dos containers (Ctrl+C para sair)
logs:
	docker compose logs -f

# Status dos containers
ps:
	docker compose ps

# ----------------------------------------------
# Testes
# ----------------------------------------------

test: ## Roda todos os testes do ms_auth
	cd $(SERVICE) && go test ./...

test-verbose: ## Testes com saída detalhada
	cd $(SERVICE) && go test -v ./...

test-verbose-auth: ## Testes detalhados do pacote de features
	cd $(SERVICE) && go test -v $(TEST_PKG)

test-cover: ## Testes com cobertura
	cd $(SERVICE) && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

test-service: ## Testa um serviço específico (SERVICE=ms_auth)
	cd $(SERVICE) && go test ./...

# ----------------------------------------------
# Build
# ----------------------------------------------

build: ## Compila o serviço
	cd $(SERVICE) && go build ./cmd/api