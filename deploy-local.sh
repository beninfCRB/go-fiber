#!/bin/bash
set -e

# Nama pod dan kontainer
POD_NAME="fiber-pod"
DB_CONTAINER="fiber-db"
APP_CONTAINER="fiber-app"
IMAGE_NAME="localhost/fiber-app:latest"

echo "=== Memulai Deployment Lokal dengan Podman ==="

# 1. Hentikan dan hapus pod/kontainer lama jika ada
if podman pod exists "$POD_NAME"; then
    echo "Menghentikan dan menghapus pod lama..."
    podman pod rm -f "$POD_NAME"
fi

# 2. Buat folder lokal untuk volume jika belum ada
mkdir -p ./keys ./storage/logs

# 3. Buat Podman Pod baru
# Port 8080 diekspos untuk aplikasi Go Fiber
# Port 5432 diekspos jika ingin mengakses database dari host
echo "Membuat Podman Pod baru: $POD_NAME..."
podman pod create \
    --name "$POD_NAME" \
    -p 8080:8080 \
    -p 5432:5432

# 4. Jalankan PostgreSQL di dalam Pod
# Karena berada di dalam pod yang sama, aplikasi Go Fiber dapat menghubunginya via localhost:5432
echo "Menjalankan PostgreSQL..."
podman run -d \
    --pod "$POD_NAME" \
    --name "$DB_CONTAINER" \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=secret \
    -e POSTGRES_DB=go_fiber_db \
    -v pgdata:/var/lib/postgresql/data \
    postgres:16-alpine

# Tunggu database siap menerima koneksi
echo "Menunggu database siap..."
until podman exec "$DB_CONTAINER" pg_isready -U postgres -d go_fiber_db >/dev/null 2>&1; do
    sleep 1
done
echo "Database siap!"

# 5. Bangun citra kontainer aplikasi
echo "Membangun citra kontainer aplikasi..."
podman build -t "$IMAGE_NAME" -f Dockerfile .

# 6. Jalankan aplikasi Go Fiber di dalam Pod
# Note: Menggunakan :z flag untuk memberikan izin SELinux yang tepat pada volume lokal
echo "Menjalankan aplikasi Go Fiber..."
podman run -d \
    --pod "$POD_NAME" \
    --name "$APP_CONTAINER" \
    -e PORT=8080 \
    -e APP_URL="http://localhost:8080" \
    -e DB_DSN="host=localhost user=postgres password=secret dbname=go_fiber_db port=5432 sslmode=disable TimeZone=Asia/Jakarta" \
    -e JWT_PRIVATE_KEY_PATH="keys/jwt_private.pem" \
    -e JWT_PUBLIC_KEY_PATH="keys/jwt_public.pem" \
    -v ./keys:/app/keys:z \
    -v ./storage:/app/storage:z \
    "$IMAGE_NAME"

echo "=== Deployment Berhasil! ==="
echo "Aplikasi berjalan di: http://localhost:8080"
echo "Untuk melihat log aplikasi: podman logs -f $APP_CONTAINER"
