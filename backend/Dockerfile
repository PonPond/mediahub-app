# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/server ./cmd/main.go

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19

# Security: run as non-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/server"]
