# ==========================================
# Tahap 1: Frontend Builder (Node.js)
# ==========================================
# Kita gunakan image Node resmi yang ringan
FROM node:lts-alpine AS frontend-builder

# Tentukan direktori kerja khusus frontend
WORKDIR /app/frontend

# Salin file package.json untuk caching dependensi
COPY frontend/package*.json ./

# Instal dependensi Vue/Pinia (NPM)
RUN npm install

# Salin seluruh kode source frontend
COPY frontend/ .

# Bangun aset produksi (Menghasilkan folder frontend/dist)
RUN npm run build


# ==========================================
# Tahap 2: Backend Builder (Golang)
# ==========================================
# (Hampir sama seperti sebelumnya, tapi versinya sudah kita sesuaikan ke 1.25)
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

# Salin dependensi modul Go
COPY go.mod go.sum ./
RUN go mod download

# Salin seluruh kode source backend
COPY . .

# Build aplikasi menjadi binary statis 'main'
RUN CGO_ENABLED=0 GOOS=linux go build -o main .


# ==========================================
# Tahap 3: Runner (Container Final)
# ==========================================
# Image Alpine murni yang super kecil
FROM alpine:latest

# Instal sertifikat dan tzdata untuk keamanan dan akurasi waktu Makassar
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 1. 🌟 KUNCI PEMBARUAN: Salin folder 'dist' hasil build Vue dari Tahap 1
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# 2. Salin binary aplikasi dari Tahap 2
COPY --from=backend-builder /app/main .

# 3. Salin folder web (template HTML) yang kita bahas sebelumnya
COPY --from=backend-builder /app/web ./web

EXPOSE 8080

CMD ["./main"]