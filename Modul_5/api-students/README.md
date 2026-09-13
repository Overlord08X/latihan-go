# api-students — Modul 5: Authentication & Security

API mahasiswa dengan sistem **autentikasi JWT** menggunakan **Go + Fiber v2 + PostgreSQL**.  
Seluruh endpoint `/students` dan `/nilais` dilindungi token — hanya user yang sudah login yang bisa mengaksesnya.  
Seluruh aplikasi (Go + PostgreSQL) berjalan di dalam **Docker container**.

## Stack Teknologi

| Komponen | Versi |
|----------|-------|
| Go | `1.26.5` (di dalam Docker) |
| Fiber v2 | `v2.52.15` |
| golang-jwt/jwt | `v5.3.1` |
| bcrypt | `golang.org/x/crypto` |
| PostgreSQL | `16-alpine` (Docker image) |
| pgx/v5 | `v5.10.0` |
| godotenv | `v1.5.1` |
| Port App | `3000` |
| Port DB | `5432` |

## Cara Menjalankan (Docker Compose)

> **Prasyarat:** Docker & Docker Compose sudah terinstall. Go **tidak** perlu diinstall di host.

```bash
# 1. Masuk ke direktori proyek
cd Modul_5/api-students

# 2. Build image dan jalankan semua container (app + PostgreSQL)
docker compose up --build -d

# 3. Cek status container
docker compose ps

# 4. Lihat log aplikasi
docker compose logs app -f

# 5. Hentikan semua container
docker compose down
```

Aplikasi otomatis:
- Menjalankan PostgreSQL dan menunggu hingga *healthy*
- Menjalankan semua migration SQL (`migrations/001` s.d. `004`)
- Menyalakan server di `http://localhost:3000`

## Konfigurasi

Salin `.env.example` ke `.env` lalu sesuaikan isinya:

```bash
cp .env.example .env
```

```env
APP_PORT=3000
DB_DSN=postgres://postgres:PASSWORD@localhost:5432/praktikum_backend?sslmode=disable

# WAJIB: minimal 32 karakter acak
# Generate: openssl rand -hex 32
JWT_SECRET=GANTI_DENGAN_RANDOM_STRING_MINIMAL_32_KARAKTER
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_DAYS=7
```

> **Catatan:** File `.env` **tidak di-commit** ke Git. Lihat `.env.example` sebagai template.

## Struktur Proyek

```text
api-students/
├── main.go                              # Entry point
├── .env.example                         # Template konfigurasi (aman di-commit)
├── .gitignore                           # .env diabaikan
├── Dockerfile                           # Multi-stage build
├── docker-compose.yml                   # Orchestrasi app + PostgreSQL
├── jwtutil/
│   └── jwt.go                           # Generate/parse JWT, hash token, MustLoadJWTSecret
├── middleware/
│   ├── auth.go                          # RequireAuth — validasi Bearer token
│   └── json.go                          # RequireJSON — validasi Content-Type
├── config/
│   ├── app.go                           # Rakit Fiber, middleware, routes, rate limiter
│   ├── env.go                           # Loader .env
│   └── logger.go                        # Structured JSON logger
├── database/
│   └── postgres.go                      # Connection pool & migration runner
├── migrations/
│   ├── 001_create_students.sql
│   ├── 002_create_nilais.sql
│   ├── 003_create_users.sql             # Tabel users (password = bcrypt hash)
│   └── 004_create_refresh_tokens.sql    # Tabel refresh_tokens (SHA-256 hash)
├── helper/
│   ├── response.go                      # Success / Fail / Created helper
│   └── query.go                         # Query string helper
├── route/
│   ├── auth.go                          # /auth/* endpoints
│   ├── student.go                       # /students/* (dilindungi RequireAuth)
│   └── nilai.go                         # /nilais/* (dilindungi RequireAuth)
└── app/
    ├── model/
    │   ├── user.go                      # User struct (password json:"-")
    │   ├── student.go
    │   ├── nilai.go
    │   └── request.go                   # RegisterRequest (tanpa field role)
    ├── repository/
    │   ├── user_repository.go
    │   ├── token_repository.go          # Simpan/rotasi/revoke refresh token
    │   ├── student_repository.go
    │   └── nilai_repository.go
    └── service/
        ├── auth_rules.go                # Validasi password (pure function, tanpa Fiber)
        ├── auth_rules_test.go           # Unit test — 9 kasus PASS
        ├── auth_service.go              # Register, Login, Refresh, Logout, Me
        ├── student_service.go
        ├── student_rules.go
        └── nilai_service.go
```

## Arsitektur: Clean Architecture + JWT Auth

```
HTTP Request
    ↓
[Middleware: RequireAuth] → parse & validasi JWT
    ↓
[Service / Handler]
    ↓
[Business Rules] → validasi murni (pure function)
    ↓
[Repository] → SQL via pgxpool
    ↓
[PostgreSQL]
```

## Kontrak API

### Autentikasi

Semua endpoint protected menggunakan header:
```
Authorization: Bearer <access_token>
```

---

### Auth Endpoints (publik)

| Metode | Endpoint | Keterangan |
|--------|----------|------------|
| `POST` | `/api/v1/auth/register` | Daftar user baru (role selalu `user`) |
| `POST` | `/api/v1/auth/login` | Login → dapat `access_token` + `refresh_token` |
| `POST` | `/api/v1/auth/refresh` | Perbarui token (rotasi refresh token) |
| `POST` | `/api/v1/auth/logout` | Cabut refresh token |
| `GET`  | `/api/v1/auth/me` | Profil user yang sedang login *(butuh token)* |

**Contoh body register:**
```json
{
  "username": "sari",
  "email": "sari@example.com",
  "password": "K0piSusu!"
}
```

**Contoh response login (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGci...",
    "refresh_token": "e029181164d3...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

---

### Students Endpoints *(semua butuh token)*

| Metode | Endpoint | Keterangan |
|--------|----------|------------|
| `GET` | `/api/v1/students` | List mahasiswa (pagination, search, sort) |
| `GET` | `/api/v1/students/:id` | Detail mahasiswa |
| `POST` | `/api/v1/students` | Tambah mahasiswa |
| `PUT` | `/api/v1/students/:id` | Ganti seluruh data |
| `PATCH` | `/api/v1/students/:id` | Ubah sebagian data |
| `DELETE` | `/api/v1/students/:id` | Hapus mahasiswa |

---

### Nilai Endpoints *(semua butuh token)*

| Metode | Endpoint | Keterangan |
|--------|----------|------------|
| `POST` | `/api/v1/nilais` | Tambah nilai mahasiswa |
| `GET` | `/api/v1/students/:nim/nilais` | Daftar nilai berdasarkan NIM |

---

### Health Check (publik)

| Metode | Endpoint | Keterangan |
|--------|----------|------------|
| `GET` | `/api/v1/health` | Cek status server (tidak butuh token) |

---

## Ringkasan Status HTTP

| Status | Situasi |
|--------|---------|
| 200 | Berhasil |
| 201 | Dibuat (register, tambah data) |
| 204 | Dihapus (tanpa body) |
| 400 | Body JSON tidak valid |
| 401 | Token tidak ada / tidak valid / kedaluwarsa |
| 409 | Data duplikat (NIM/username/email sudah ada) |
| 415 | `Content-Type` bukan `application/json` |
| 422 | Validasi gagal, rincian per field |
| 429 | Terlalu banyak percobaan login (rate limiter) |

## Keamanan yang Diimplementasikan

| Ancaman | Cara Menutup |
|---------|-------------|
| Algorithm confusion | `keyfunc` hanya izinkan HS256 secara eksplisit |
| Secret lemah | `MustLoadJWTSecret` — panic jika < 32 karakter |
| Brute force | Rate limiter: maks 5 req/30 detik → `429 + Retry-After` |
| User enumeration | Pesan error **sama persis** untuk salah password & username tidak ada |
| Mass assignment | Struct `RegisterRequest` tidak punya field `role` |
| Password bocor | Tag `json:"-"` pada field `Password` di struct `User` |
| Token tidak bisa dicabut | Refresh token berumur pendek + dapat di-revoke |
| Payload besar | `BodyLimit: 1MB` di konfigurasi Fiber |

## Penggunaan AI

Bantuan AI (Google Antigravity) digunakan sebagai pendamping pembelajaran untuk:

- Memahami konsep **JWT** — anatomi header, payload, signature, dan cara verifikasinya.
- Memahami cara menutup kerentanan **algorithm confusion** dengan memvalidasi signing method secara eksplisit.
- Memahami pola **timing-safe comparison** saat login untuk mencegah *user enumeration*.
- Memahami mekanisme **rotasi refresh token** dan alasan disimpan sebagai hash SHA-256.
- Membantu penulisan **fungsi murni** (`ValidatePasswordStrength`) yang dapat diuji secara unit tanpa Fiber.
- Mengkonfigurasi **rate limiter** Fiber sebagai middleware per-route.
- Membantu memeriksa kode dan menjelaskan error build, termasuk masalah *import cycle*.
