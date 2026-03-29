# --- Stage 1: Build Backend (Golang) ---
FROM golang:alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod dan sum dulu untuk caching layer
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code backend (internal, cmd, pkg, dll)
COPY . .
# Copy package.json dan package-lock.json frontend untuk install dependencies
# COPY frontend/package*.json ./frontend/
# Install dependencies dan build frontend
# RUN cd frontend && npm install && npm run build

# Build aplikasi Golang (sesuaikan dengan lokasi main.go Anda)
RUN go build -o main ./cmd/api/main.go

# --- Stage 2: Runtime (Final Image) ---
FROM alpine:latest

WORKDIR /app

# Install dependencies yang dibutuhkan alpine (opsional)
RUN apk add --no-cache ca-certificates tzdata

# Copy binary dari builder
COPY --from=builder /app/main .

# Copy folder statis yang dibutuhkan (Templates, Assets, Configs)
COPY --from=builder /app/web ./web
# COPY --from=builder /app/configs* ./configs/

# --- CATATAN PENTING ---
# Bagian COPY frontend di bawah ini kita komentari dulu agar 
# Railway tidak error saat mencoba mencari folder frontend/dist
# COPY --from=builder /app/frontend/dist ./frontend/dist

# Expose port (sesuaikan dengan port Gin Anda, misal 8080)
EXPOSE 8080

# Jalankan aplikasi
CMD ["./main"]
