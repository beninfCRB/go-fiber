# Go Fiber v3 Authentication & User Management Service

Layanan RESTful API terpusat untuk Autentikasi dan Manajemen Pengguna yang dibangun menggunakan **Go**, **Fiber v3**, **GORM (PostgreSQL)**, dan pengamanan token asimetris **JWT RS256**. 

Layanan ini dirancang dengan prinsip *Clean Code* (pemisahan Handler, Service, Repository, DTO) dan siap diintegrasikan sebagai *Central Auth Service* pada arsitektur Microservices.

---

## Fitur Utama

1. **Autentikasi Asimetris (JWT RS256)**:
   * Menggunakan algoritma asimetris `RS256` untuk keamanan tingkat tinggi.
   * Pasangan kunci privat (`jwt_private.pem`) dan publik (`jwt_public.pem`) **otomatis dibuat saat aplikasi dinyalakan pertama kali** jika belum tersedia di folder `keys/`.
   * Layanan microservice lain hanya memerlukan kunci publik untuk memvalidasi akses token pengguna secara terdesentralisasi.
2. **Kinerja Database Optimal (PostgreSQL & GORM)**:
   * Kolom pencarian penting seperti `email`, `verification_token`, dan `reset_token` telah diberi **Indeks Database** untuk memastikan kecepatan pencarian kueri meskipun volume data pengguna bertambah besar.
3. **Tata Kelola Keamanan (Audit Logging)**:
   * Pencatatan log aktivitas administratif penting secara otomatis ke dalam database (seperti peristiwa login sukses/gagal, registrasi, pergantian peran, perubahan status aktif akun, dan CRUD menu).
   * Kolom `action` dan `created_at` pada log audit telah dilengkapi indeks untuk analisis audit yang cepat.
4. **Proteksi Akses (Rate Limiting)**:
   * Melindungi endpoint sensitif (`/auth/register`, `/auth/login`, `/auth/forgot-password`, `/auth/reset-password`) dari serangan *brute force* dengan batasan maksimal **5 request per 1 menit** untuk setiap IP Address.
5. **Verifikasi Email & Pemulihan Kata Sandi**:
   * Alur pendaftaran terverifikasi menggunakan token verifikasi email.
   * Alur pemulihan kata sandi menggunakan token sementara dengan batas kadaluwarsa 1 jam.
   * Mendukung pengiriman asli menggunakan **SMTP** produksi atau **Simulasi Fallback** yang mencatat isi email secara rapi di file lokal `storage/logs/mails.log` (sangat berguna untuk tahap development).
6. **Manajemen Hak Akses Menu Dinamis**:
   * Struktur menu bertingkat (*tree structure*) yang dikaitkan langsung dengan otorisasi peran pengguna (*Role-based Access Control - RBAC*).

---

## Prasyarat Sistem

* **Go**: Versi 1.22 atau lebih baru.
* **PostgreSQL**: Database relasional terpasang dan aktif.

---

## Cara Penggunaan (Panduan Mulai Cepat)

### Langkah 1: Kloning & Pengaturan Lingkungan
Salin berkas template `.env.example` menjadi `.env` di direktori utama proyek:
```bash
cp .env.example .env
```
Sesuaikan konfigurasi di dalam berkas `.env` (terutama parameter koneksi PostgreSQL `DB_DSN` dan konfigurasi SMTP jika ingin menggunakan email asli).

### Langkah 2: Setup Database PostgreSQL
Buatlah database kosong di PostgreSQL Anda. Contoh nama database: `go_fiber_db`.
GORM akan secara otomatis membuat dan melakukan migrasi tabel (`AutoMigrate`) pada saat aplikasi dijalankan, meliputi tabel:
* `users`
* `role_models`
* `user_roles`
* `refresh_tokens`
* `menus`
* `menu_roles`
* `audit_logs`

### Langkah 3: Menjalankan Aplikasi
Jalankan aplikasi menggunakan perintah berikut:
```bash
go run cmd/main.go
```
*Saat pertama kali dinyalakan, aplikasi akan secara otomatis mendeteksi bahwa kunci RSA belum ada, lalu membuat folder `keys/` dan berkas `jwt_private.pem` serta `jwt_public.pem` secara dinamis.*

Aplikasi akan berjalan pada port default: `http://localhost:8080`.

---

## Cara Melakukan Pengujian Alur Registrasi & Verifikasi (Pengembangan Lokal)

1. **Registrasi Pengguna**:
   Kirim permintaan `POST` ke `/auth/register` dengan payload nama, email, dan sandi.
2. **Ambil Token Verifikasi**:
   Buka file log simulasi email di `storage/logs/mails.log`. Ambil tautan verifikasi yang tercatat di sana, contoh:
   `http://localhost:8080/auth/verify-email?token=abcdef...`
3. **Verifikasi Email**:
   Kirim permintaan `GET` ke URL verifikasi di atas (dapat dibuka langsung di browser) untuk mengaktifkan akun.
4. **Login**:
   Setelah email terverifikasi, kirim permintaan `POST` ke `/auth/login` untuk mendapatkan token JWT asimetris (`RS256`) beserta `refresh_token` untuk rotasi sesi.

---

## Daftar Endpoint API Utama

| Metode | Endpoint | Proteksi | Hak Akses | Deskripsi |
| :--- | :--- | :--- | :--- | :--- |
| **POST** | `/auth/register` | Publik (Rate Limited) | Semua | Registrasi pengguna baru (default role: `user`) |
| **POST** | `/auth/login` | Publik (Rate Limited) | Semua | Masuk sistem dan mendapatkan token JWT RS256 |
| **GET** | `/auth/verify-email` | Publik | Semua | Memverifikasi akun pengguna via token email |
| **POST** | `/auth/forgot-password` | Publik (Rate Limited) | Semua | Mengirim tautan pemulihan kata sandi |
| **POST** | `/auth/reset-password` | Publik (Rate Limited) | Semua | Memperbarui kata sandi baru menggunakan token |
| **POST** | `/auth/refresh` | Publik | Semua | Memperbarui Access Token menggunakan Refresh Token |
| **POST** | `/auth/logout` | Publik | Semua | Keluar dari sesi aktif saat ini |
| **GET** | `/api/profile` | JWT (`RS256`) | Semua | Melihat profil detail pengguna yang masuk |
| **POST** | `/api/logout-all` | JWT (`RS256`) | Semua | Keluar dari seluruh sesi perangkat aktif |
| **GET** | `/api/menu` | JWT (`RS256`) | Semua | Mengambil daftar sidebar menu sesuai otorisasi peran |
| **GET** | `/api/admin/users` | JWT (`RS256`) | Admin, Super Admin | Mengambil daftar seluruh pengguna terpaginasi |
| **PATCH** | `/api/admin/users/:id/active` | JWT (`RS256`) | Admin, Super Admin | Mengaktifkan/menonaktifkan akun pengguna |
| **PATCH** | `/api/super-admin/users/:id/role` | JWT (`RS256`) | Super Admin | Mengubah peran pengguna (Assign Role) |
| **POST** | `/api/super-admin/menus` | JWT (`RS256`) | Super Admin | CRUD / Manajemen struktur menu dinamis |

---

## Dokumentasi API (Swagger)

Aplikasi telah dilengkapi dengan Swagger UI dinamis yang dapat diakses langsung pada browser setelah server dijalankan:
* **Swagger UI**: [http://localhost:8080/swagger](http://localhost:8080/swagger)
* **Swagger JSON**: [http://localhost:8080/swagger.json](http://localhost:8080/swagger.json)
