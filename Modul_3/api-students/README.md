# api-students — Modul 3: Database & Repository Pattern

API mahasiswa menggunakan **Go + Fiber v2 + PostgreSQL** dengan arsitektur Repository Pattern.

## Environment

- Go `1.26.5`
- Fiber v2 `v2.52.15`
- PostgreSQL `18.4`
- pgx/v5 `v5.10.0`
- godotenv `v1.5.1`
- Port: `3000` (dapat dikonfigurasi via `.env`)

## Persiapan Database

```sql
-- Buat database
CREATE DATABASE praktikum_backend;
```

Migration tabel dijalankan otomatis saat server pertama kali dinyalakan.

## Konfigurasi

Salin dan sesuaikan file `.env`:

```env
APP_PORT=3000
DB_DSN=postgres://postgres:password@localhost:5432/praktikum_backend?sslmode=disable
```

## Cara Menjalankan

```powershell
go run .
```

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
