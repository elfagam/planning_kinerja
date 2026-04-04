# --- Stage 1: Build Frontend (Vite/Vue) ---
FROM node:20-alpine3.20 AS frontend-builder
WORKDIR /app/frontend
# Copy package.json and package-lock.json first for caching
COPY frontend/package*.json ./
RUN npm install
# Copy frontend source and build
COPY frontend/ ./
RUN npm run build

# --- Stage 2: Build Backend (Go) ---
FROM golang:1.23-alpine3.20 AS backend-builder
WORKDIR /app
# Copy go mod and sum first for caching
COPY go.mod go.sum ./
RUN go mod download
# Copy backend source
COPY . .
# Build Golang application (API & Migration)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate/main.go

# --- Stage 3: Runtime (Final Image) ---
FROM alpine:3.20
WORKDIR /app

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Copy binaries from backend-builder
COPY --from=backend-builder --chown=appuser:appgroup /app/main .
COPY --from=backend-builder --chown=appuser:appgroup /app/migrate .

# Copy static folders (Templates & Assets)
COPY --from=backend-builder --chown=appuser:appgroup /app/web ./web

# Copy migrations folder
COPY --from=backend-builder --chown=appuser:appgroup /app/migrations ./migrations

# Copy startup script
COPY --from=backend-builder --chown=appuser:appgroup /app/scripts/entrypoint.sh ./scripts/entrypoint.sh

# Copy frontend build output from frontend-builder
COPY --from=frontend-builder --chown=appuser:appgroup /app/frontend/dist ./frontend/dist

# Ensure entrypoint is executable
RUN chmod +x ./scripts/entrypoint.sh

# Expose port (default 8080)
EXPOSE 8080

# Run via entrypoint script
ENTRYPOINT ["./scripts/entrypoint.sh"]
