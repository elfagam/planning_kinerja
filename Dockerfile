# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Set build arguments for static binary
ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o e-plan-ai ./cmd/api
RUN go build -ldflags="-s -w" -o gorm-migrate ./cmd/gorm-migrate

# Stage 2: Runtime
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user for security
RUN adduser -D -u 1001 eplan
USER eplan

WORKDIR /app
COPY --from=builder --chown=eplan:eplan /app/e-plan-ai .
COPY --from=builder --chown=eplan:eplan /app/gorm-migrate .
COPY --from=builder --chown=eplan:eplan /app/web ./web
COPY --from=builder --chown=eplan:eplan /app/migrations ./migrations
COPY --from=builder --chown=eplan:eplan /app/docs ./docs

# Set timezone
ENV TZ=Asia/Jakarta

EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["sh", "-c", "./gorm-migrate || true; ./e-plan-ai"]
