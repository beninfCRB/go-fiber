#!/bin/bash
set -e

# Konfigurasi
COMPOSE_FILE="docker-compose.yml"
APP_CONTAINER="fiber-app"
SMTP_HOST="smtp.gmail.com"
SMTP_PORT=587
SMTP_USER="beninf10@gmail.com"
SMTP_PASS="wlkt gaeb ezbh dkqj"
SMTP_FROM="beninf10@gmail.com"

echo "=== Memulai Deployment Lokal dengan Docker ==="

# 1. Buat folder lokal untuk volume jika belum ada
mkdir -p ./keys ./storage/logs

# 2. Export variabel SMTP agar bisa dipakai docker-compose
export SMTP_HOST SMTP_PORT SMTP_USER SMTP_PASS SMTP_FROM

# 3. Hentikan dan hapus kontainer lama jika ada
echo "Menghentikan kontainer lama (jika ada)..."
docker compose -f "$COMPOSE_FILE" down --remove-orphans

# 4. Build ulang image aplikasi
echo "Membangun image kontainer aplikasi..."
docker compose -f "$COMPOSE_FILE" build --no-cache app

# 5. Jalankan semua service
echo "Menjalankan semua service..."
docker compose -f "$COMPOSE_FILE" up -d

# 6. Tunggu aplikasi siap
echo "Menunggu aplikasi siap..."
sleep 3
docker compose -f "$COMPOSE_FILE" ps

echo ""
echo "=== Deployment Berhasil! ==="
echo "Aplikasi berjalan di: http://localhost:8081"
echo "Untuk melihat log aplikasi: docker logs -f $APP_CONTAINER"
echo "Untuk menghentikan: docker compose down"
