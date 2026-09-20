# api-students — Modul 6: Authorization & Role Based Access Control (RBAC)

API mahasiswa dengan sistem **Autentikasi JWT** dan **Otorisasi Berbasis Peran (RBAC) + Kepemilikan Data (Ownership)** menggunakan **Go + Fiber v2 + PostgreSQL**.  
Seluruh endpoint dilindungi sesuai matriks hak akses peran (`admin`, `staff`, `user`) dan aturan kepemilikan data.  
Seluruh aplikasi berjalan di dalam **Docker container**.

---

## Stack Teknologi

| Komponen | Versi |
|----------|-------|
| Go | `1.24+` (di dalam Docker) |
| Fiber v2 | `v2.52.15` |
| golang-jwt/jwt | `v5.3.1` |
| bcrypt | `golang.org/x/crypto` |
| PostgreSQL | `16-alpine` (Docker image) |
| pgx/v5 | `v5.10.0` |
| godotenv | `v1.5.1` |
| Port App | `3000` |
| Port DB | `5432` |

---

## Cara Menjalankan (Docker Compose)

> **Prasyarat:** Docker & Docker Compose sudah terinstall.

```bash
# 1. Masuk ke direktori proyek
cd Modul_6/api-students

# 2. Build image dan jalankan container
docker compose up --build -d

# 3. Cek status container
docker compose ps

# 4. Lihat log aplikasi
docker compose logs app -f

# 5. Hentikan container
docker compose down
```

Aplikasi otomatis:
- Menjalankan PostgreSQL dan menunggu hingga *healthy*
- Menjalankan seluruh migration SQL (`migrations/001` s.d. `006`)
- Memuat permissions RBAC ke memori saat startup (*fail-closed*)
- Menyalakan server di `http://localhost:3000`

---

## Struktur Arsitektur & Otorisasi

```text
HTTP Request
    ↓
[Middleware: RequireAuth] → Verifikasi identitas (JWT)
    ↓
[Middleware: RequirePermission] → Otorisasi statis level rute (user:list, student:create, dll.)
    ↓
[Service / Handler]
    ↓
[Pure Rules: CanAccessStudent / CanAccessUser] → Otorisasi dinamis level data (Ownership)
    ↓
[Repository] → SQL via pgxpool
    ↓
[PostgreSQL 16]
```

---

## Matriks Hak Akses

### 1. Users Endpoints
| Metode | Endpoint | Guard | Hak Akses |
|---|---|---|---|
| `GET` | `/api/v1/users` | `RequirePermission` | `admin`, `staff` (200), `user` (403) |
| `GET` | `/api/v1/users/:id` | Service (`CanAccessUser`) | Diri sendiri (200), orang lain butuh `user:read:any` |
| `PATCH` | `/api/v1/users/:id/role` | `RequirePermission` + Rule | `admin` orang lain (200), ubah diri sendiri (422) |
| `DELETE` | `/api/v1/users/:id` | `RequirePermission` + Rule | `admin` orang lain (204), hapus diri sendiri (403) |

### 2. Students Endpoints
| Metode | Endpoint | Guard | Hak Akses |
|---|---|---|---|
| `GET` | `/api/v1/students` | `RequirePermission` | `admin`, `staff` (200), `user` (403) |
| `POST` | `/api/v1/students` | `RequirePermission` | `admin`, `staff` (201), `user` (403) |
| `GET` | `/api/v1/students/:id` | Service (`CanAccessStudent`) | Pemilik data (200), orang lain butuh `student:read:any` |
| `PUT` | `/api/v1/students/:id` | Service (`CanAccessStudent`) | Pemilik data (200), orang lain butuh `student:update:any` |
| `PATCH` | `/api/v1/students/:id` | Service (`CanAccessStudent`) | Pemilik data (200), orang lain butuh `student:update:any` |
| `DELETE` | `/api/v1/students/:id` | `RequirePermission` | `admin` (204), `staff`/`user` (403) |

---

## Pengujian Unit

Jalankan pengujian unit tanpa database:
```bash
docker run --rm -v "$PWD":/app -w /app golang:alpine go test -v ./app/service/...
```
Seluruh 40 unit test mencakup validasi password, registrasi, otorisasi RBAC pengguna, otorisasi kepemilikan mahasiswa, dan validasi data berstatus **PASS**.
