FROM golang:1.27.0-alpine AS builder

WORKDIR /app

COPY shared/    ./shared/
COPY ms_auth/   ./ms_auth/

RUN printf 'go 1.27.0\n\nuse (\n    ./shared\n    ./ms_auth\n)\n' > go.work

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd ms_auth && go work sync && go mod download

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd ms_auth && CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/api ./cmd/api

FROM alpine:3.19
WORKDIR /root/
COPY --from=builder /app/bin/api ./ms_auth
RUN chmod +x /root/ms_auth && apk --no-cache add tzdata
ENV SERVER_PORT=4001
ENTRYPOINT ["/root/ms_auth"]