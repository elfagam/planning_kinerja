# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o e-plan-ai ./cmd/api
RUN go build -o gorm-migrate ./cmd/gorm-migrate

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/e-plan-ai .
COPY --from=builder /app/gorm-migrate .
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/.env.example .env

EXPOSE 8080
CMD ["sh", "-c", "./gorm-migrate || true; ./e-plan-ai"]
