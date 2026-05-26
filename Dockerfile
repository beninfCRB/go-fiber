# Stage 1: Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install git, ca-certificates, dan tzdata
RUN apk add --no-cache git ca-certificates tzdata

# Salin file modul
COPY go.mod go.sum ./

# Unduh dependensi
RUN go mod download

# Salin seluruh kode sumber
COPY . .

# Kompilasi aplikasi secara statis
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/main.go

# Stage 2: Runtime
FROM alpine:latest AS runner

WORKDIR /app

# Salin sertifikat SSL dan zona waktu dari builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Salin binary aplikasi
COPY --from=builder /app/main .

# Salin file .env.example sebagai referensi default
COPY --from=builder /app/.env.example .env.example

# Buat direktori untuk kunci JWT dan storage logs
RUN mkdir -p keys storage/logs

# Port aplikasi
EXPOSE 8081

# Jalankan aplikasi
CMD ["./main"]
