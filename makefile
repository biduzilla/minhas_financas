# Makefile para testes em monorepo Go (Windows)

.PHONY: test test-verbose test-cover test-service

TEST_PKG ?= ./internal/features/user

# Alvo padrão: roda testes do ms_auth
test:
	cd ms_auth && go test ./...

# Testes com saída detalhada
test-verbose:
	cd ms_auth && go test -v ./...

test-verbose-user:
	cd ms_auth && go test -v $(TEST_PKG)

# Testes com cobertura
test-cover:
	cd ms_auth && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

# Testa um serviço específico (use: make test-service SERVICE=ms_auth)
test-service:
	cd $(SERVICE) && go test ./...