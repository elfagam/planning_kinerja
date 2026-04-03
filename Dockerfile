# --- Stage 1: Build Frontend (Vite/Vue) ---
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
# Copy package.json dan package-lock.json dulu untuk caching
COPY frontend/package*.json ./
RUN npm install
# Copy seluruh source code frontend dan build
COPY frontend/ ./
RUN npm run build

# --- Stage 2: Build Backend (Go) ---
FROM golang:alpine AS backend-builder
WORKDIR /app
# Copy go mod dan sum dulu untuk caching layer
COPY go.mod go.sum ./
RUN go mod download
# Copy seluruh source code backend
COPY . .
# Build aplikasi Golang (API & Migrasi)
RUN go build -o main ./cmd/api/main.go
RUN go build -o migrate ./cmd/migrate/main.go

# --- Stage 3: Runtime (Final Image) ---
FROM alpine:latest
WORKDIR /app
# Install dependencies yang dibutuhkan alpine
RUN apk add --no-cache ca-certificates tzdata

# Copy binary dari backend-builder
COPY --from=backend-builder /app/main .
COPY --from=backend-builder /app/migrate .
# Copy folder statis (Templates & Assets)
COPY --from=backend-builder /app/web ./web
# Copy folder migrations
COPY --from=backend-builder /app/migrations ./migrations
# Copy startup script
COPY --from=backend-builder /app/scripts/entrypoint.sh ./scripts/entrypoint.sh
# Copy hasil build frontend dari frontend-builder
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Pastikan entrypoint dapat dijalankan
RUN chmod +x ./scripts/entrypoint.sh

# Expose port (default internal 8080)
EXPOSE 8080

# Jalankan aplikasi melalui entrypoint script
ENTRYPOINT ["./scripts/entrypoint.sh"]
