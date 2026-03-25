# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o e-plan-ai ./cmd/api

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/e-plan-ai .
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/.env.example .env

EXPOSE 8080
CMD ["./e-plan-ai"]
