# services/notification/Dockerfile

# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy module files first so this layer is cached until deps change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source tree (all services share one module root).
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
      -ldflags="-s -w" \
      -o /notification-service \
      ./services/notification

# ── Stage 2: run ─────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /notification-service .

EXPOSE 8005

ENTRYPOINT ["/app/notification-service"]
