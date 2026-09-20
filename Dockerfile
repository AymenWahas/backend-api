# =========================
# Stage 1: Builder
# =========================
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o backend-api ./cmd/api


# =========================
# Stage 2: Runtime
# =========================
FROM alpine:latest
WORKDIR /app

RUN adduser -D -H appuser

COPY --from=builder /app/backend-api ./backend-api
COPY certs ./certs

RUN chown -R appuser:appuser /app/certs

USER appuser

CMD ["./backend-api"]