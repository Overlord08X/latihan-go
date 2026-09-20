# ==============================================================================
# Stage 1 — Build
# Menggunakan image resmi Go Alpine untuk menghasilkan binary yang ringan.
# ==============================================================================
FROM golang:alpine AS builder

# Install dependensi build (git diperlukan untuk go mod download beberapa paket)
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod dan go.sum terlebih dahulu agar layer di-cache saat source berubah
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build binary; CGO_ENABLED=0 agar tidak bergantung pada libc host (penting untuk Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -o api-students .

# ==============================================================================
# Stage 2 — Run
# Image final sangat kecil karena hanya berisi binary + folder migrations
# ==============================================================================
FROM alpine:3.20

WORKDIR /app

# Salin binary dari stage builder
COPY --from=builder /app/api-students .

# Salin folder migrations (dibutuhkan saat startup)
COPY --from=builder /app/migrations ./migrations

# Port default aplikasi
EXPOSE 3000

CMD ["./api-students"]
