# api-students — Modul 3: Database & Repository Pattern

API CRUD mahasiswa menggunakan **Go + Fiber v2 + PostgreSQL** dengan arsitektur Repository Pattern.  
Seluruh aplikasi (Go + PostgreSQL) berjalan di dalam **Docker container** — tidak perlu install Go di sistem host.

## Stack Teknologi

| Komponen | Versi |
|----------|-------|
| Go | `1.26.5` (di dalam Docker) |
| Fiber v2 | `v2.52.15` |
| PostgreSQL | `16-alpine` (Docker image) |
| pgx/v5 | `v5.10.0` |
| godotenv | `v1.5.1` |
| Docker Compose | `v5.x` |
| Port App | `3000` |
| Port DB | `5432` |

## Cara Menjalankan (Docker Compose)

> **Prasyarat:** Docker & Docker Compose sudah terinstall. Go **tidak** perlu diinstall di host.

```bash
# 1. Clone / masuk ke direktori proyek
cd Modul_3/api-students

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
- Menjalankan migration SQL (`migrations/001_create_students.sql`)
- Menyalakan server di `http://localhost:3000`

## Konfigurasi (Opsional)

File `.env` digunakan saat menjalankan **tanpa Docker** (local development):

```env
APP_PORT=3000
DB_DSN=postgres://postgres:171005@localhost:5432/praktikum_backend?sslmode=disable
```

Saat menggunakan Docker Compose, konfigurasi sudah di-set via `environment` di `docker-compose.yml`.

## Struktur Proyek

```text
api-students/
├── main.go                          # Entry point
├── model.go                         # Request bodies & response envelope
├── helper.go                        # Helper fungsi respons, validasi
├── handler.go                       # HTTP handlers (struct + DI)
├── .env                             # Konfigurasi (tidak di-commit ke Git)
├── config/
│   └── env.go                       # Loader konfigurasi
├── database/
│   └── postgres.go                  # Connection pool & migration runner
├── migrations/
│   └── 001_create_students.sql      # Schema tabel students
└── app/
    ├── model/
    │   └── student.go               # Entitas domain Student
    └── repository/
        └── student_repository.go    # Interface + implementasi PostgreSQL
```

## Arsitektur: Repository Pattern

```
HTTP Request
    ↓
[Handler] → memanggil interface
    ↓
[StudentRepository interface]
    ↓
[pgStudentRepository] → SQL via pgxpool
    ↓
[PostgreSQL]
```

Keunggulan:
- **Handler tidak tahu implementasi DB** — bisa diganti ke MongoDB tanpa ubah handler.
- **Mudah diuji** — tinggal buat mock yang mengimplementasikan interface.
- **Pemisahan concern** — HTTP logic terpisah dari data access logic.

## Kontrak API

### Struktur Respons Seragam

```json
{
  "success": true | false,
  "message": "...",
  "data": { ... } | [ ... ],
  "meta": { "page": 1, "limit": 10, "total": 3, "total_pages": 1 },
  "errors": { "field": "pesan error" }
}
```

---

### GET /api/v1/students

**Query string:**

| Parameter   | Default | Keterangan |
|-------------|---------|-----------|
| `page`      | 1       | Nomor halaman |
| `limit`     | 10      | Per halaman (maks. 100) |
| `search`    | —       | Cari nama (case-insensitive) |
| `sort`      | id      | `id`, `nim`, `name`, `grade` |
| `order`     | asc     | `asc` atau `desc` |
| `is_active` | —       | `true` atau `false` |

**Status:** `200`

---

### GET /api/v1/students/:id

**Status yang mungkin:** `200`, `400`, `404`

---

### POST /api/v1/students

**Header wajib:** `Content-Type: application/json`

**Body:**

```json
{
  "nim": "2023010",
  "name": "Dian Pratiwi",
  "grade": 88.0,
  "is_active": true
}
```

**Status yang mungkin:** `201`, `400`, `409`, `415`, `422`

Header respons: `Location: /api/v1/students/{id}`

---

### PUT /api/v1/students/:id

Mengganti **seluruh** data. Semua field wajib.

**Status yang mungkin:** `200`, `400`, `404`, `409`, `415`, `422`

---

### PATCH /api/v1/students/:id

Mengubah **sebagian** data. Hanya field yang dikirim yang berubah.

**Status yang mungkin:** `200`, `400`, `404`, `409`, `415`, `422`

---

### DELETE /api/v1/students/:id

Menghapus mahasiswa. **204 tanpa body.**

**Status yang mungkin:** `204`, `400`, `404`

---

## Ringkasan Status HTTP

| Status | Situasi |
|--------|---------|
| 200 | Pengambilan atau perubahan berhasil |
| 201 | Penambahan berhasil + header `Location` |
| 204 | Penghapusan berhasil, tanpa body |
| 400 | Body JSON tidak sah atau id bukan angka |
| 404 | Data tidak ditemukan |
| 409 | NIM sudah digunakan (unique constraint) |
| 415 | `Content-Type` bukan `application/json` |
| 422 | Validasi isi gagal, dengan rincian per field |

## Penggunaan AI

Bantuan AI digunakan sebagai pendamping pembelajaran untuk:

- Memahami pola Repository Pattern dan cara mengimplementasikan interface di Go.
- Memahami penggunaan `pgxpool` untuk connection pool yang aman untuk concurrent request.
- Memahami cara membangun query SQL dinamis dengan parameter aman (parameterized query).
- Memahami sentinel errors dan cara menggunakannya dengan `errors.Is()`.
- Memahami teknik dependency injection pada handler Go.
- Membantu memeriksa kode dan menjelaskan error.
