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
FROM golang:1.24-alpine AS backend-builder
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
FROM alpine:latest
WORKDIR /app

# Install necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binaries from backend-builder
COPY --from=backend-builder /app/main .
COPY --from=backend-builder /app/migrate .

# Copy static folders (Templates & Assets)
COPY --from=backend-builder /app/web ./web

# Copy migrations folder
COPY --from=backend-builder /app/migrations ./migrations

# Copy startup script
COPY --from=backend-builder /app/scripts/entrypoint.sh ./scripts/entrypoint.sh

# Copy frontend build output from frontend-builder
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Ensure binaries and entrypoint are executable
RUN chmod +x ./main ./migrate ./scripts/entrypoint.sh

# Create a non-root user and set ownership
RUN addgroup -S appgroup && adduser -S appuser -G appgroup && \
    chown -R appuser:appgroup /app

USER appuser

# Expose port (default 8080)
EXPOSE 8080

# Run via entrypoint script
ENTRYPOINT ["./scripts/entrypoint.sh"]
