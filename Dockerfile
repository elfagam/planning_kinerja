# Stage 1: Build Frontend (Node.js)
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Backend (Go)
FROM golang:1.24-alpine AS builder

# Set build arguments for static binary
ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o e-plan-ai ./main.go
RUN go build -ldflags="-s -w" -o gorm-migrate ./cmd/gorm-migrate

# Stage 3: Runtime
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user for security
RUN adduser -D -u 1001 eplan

WORKDIR /app
COPY --from=builder --chown=eplan:eplan /app/e-plan-ai .
COPY --from=builder --chown=eplan:eplan /app/gorm-migrate .
COPY --from=builder --chown=eplan:eplan /app/web ./web
COPY --from=builder --chown=eplan:eplan /app/migrations ./migrations
COPY --from=builder --chown=eplan:eplan /app/docs ./docs
# Copy frontend dist from builder
COPY --from=frontend-builder --chown=eplan:eplan /app/frontend/dist ./frontend/dist

USER eplan

# Set timezone
ENV TZ=Asia/Jakarta

CMD ["sh", "-c", "./gorm-migrate || true; ./e-plan-ai"]
