# api-students — Modul 7: Advanced API Design

API mahasiswa dengan implementasi standar industri **Advanced API Design** menggunakan **Go (Fiber v2) + PostgreSQL 16**.  
Proyek ini mengadopsi validasi deklaratif terstruktur, pagination berkinerja tinggi berbasis kursor, negosiasi konten ganda (JSON & streaming CSV), serta penanganan galat terpusat berbasis varian RFC 7807. Seluruh sistem berjalan di dalam **Docker container**.

---

## Fitur Utama Modul 7

1. **Validasi Deklaratif (Declarative Validation):**
   - Menggunakan pustaka `go-playground/validator/v10` pada DTO struct request.
   - Pendaftaran custom tag validator `nim` menggunakan regular expression numerik tepat 9 digit (`^\d{9}$`).
   - Tag `omitnil` untuk mutasi parsial (PATCH) yang membedakan secara presisi antara field yang dihilangkan (*omitted* / bernilai `nil`), field kosong yang tidak valid (`""`), dan field dengan nilai valid.

2. **Keyset / Cursor-Based Pagination:**
   - Menghindari kelemahan paginasi konvensional `OFFSET/LIMIT` yang lambat ($O(N)$) dan rentan terhadap anomali *data drift* saat penambahan baris baru secara simultan.
   - Paginasi deterministik $O(1)$ menggunakan kursor opaque terenkapsulasi Base64 dengan tuple komposit `(created_at, id)`:
     ```sql
     WHERE (created_at, id) < ($1, $2)
     ORDER BY created_at DESC, id DESC
     LIMIT $3
     ```
   - Dioptimasi menggunakan PostgreSQL composite index: `students_created_at_id_desc_idx` dan `users_created_at_id_desc_idx`.

3. **Negosiasi Konten (Content Negotiation):**
   - Mendukung representasi ganda pada endpoint koleksi berdasarkan header `Accept`:
     - `application/json` atau wildcard `*/*` (default) $\rightarrow$ Respons JSON standar berstruktur `data` dan `meta`.
     - `text/csv` $\rightarrow$ Streaming serialisasi CSV baris demi baris menggunakan `csv.Writer` dengan pemanggilan eksplisit `writer.Flush()` guna mencegah *memory bloat*.
   - Mengembalikan status `406 Not Acceptable` jika klien meminta representasi selain format yang didukung (misal: XML atau YAML).
   - Validasi ketat `Content-Type: application/json` pada request mutasi (POST/PUT/PATCH) dengan status `415 Unsupported Media Type` jika tidak sesuai.

4. **Standarisasi Respon Galat & Audit Logging:**
   - Centralized error handler mengadopsi varian RFC 7807 (*Problem Details for HTTP APIs*).
   - Tipe galat kustom `helper.AppError` dengan pemisahan status dan kode yang konsisten (`VALIDATION_ERROR`, `NOT_FOUND`, `UNAUTHORIZED`, `PERMISSION_DENIED`, `NOT_ACCEPTABLE`, `UNSUPPORTED_MEDIA_TYPE`, `INTERNAL_ERROR`).
   - Downstream request logger (`config/logger.go`) yang merekam status kode HTTP aktual pasca-eksekusi handler serta pemisahan level log (`WARN` untuk kesalahan klien 4xx dan `ERROR` untuk kegagalan server 5xx) guna mencegah *information disclosure*.

---

## Stack Teknologi

| Komponen | Versi / Keterangan |
|---|---|
| Bahasa Pemrograman | Go `1.24+` (Docker Container) |
| Web Framework | Fiber v2 (`v2.52.15`) |
| Validasi Input | `go-playground/validator/v10` (`v10.30.2`) |
| Database Engine | PostgreSQL 16 (`16-alpine`) |
| Database Driver | `pgx/v5` (`v5.10.0`) dengan connection pooling |
| Token Keamanan | JWT (`golang-jwt/jwt/v5` `v5.3.1`) |
| Enkripsi Sandi | bcrypt (`golang.org/x/crypto`) |
| Port Aplikasi | `3000` |
| Port PostgreSQL | `5432` |

---

## Cara Menjalankan Aplikasi (Docker Compose)

> **Prasyarat:** Docker daemon dan Docker Compose telah terpasang pada sistem.

```bash
# 1. Masuk ke direktori proyek
cd Modul_7/api-students

# 2. Hentikan container aktif (bila ada)
docker compose down

# 3. Bangun ulang image dan jalankan container di latar belakang
docker compose up --build -d

# 4. Verifikasi status kesehatan kontainer
docker compose ps

# 5. Pantau log terstruktur secara langsung
docker compose logs app -f

# 6. Hentikan seluruh layanan
docker compose down
```

Layanan otomatis:
- Menunggu status *healthy* dari PostgreSQL `praktikum_db`.
- Menjalankan 8 berkas migrasi SQL (`001` s.d. `008_cursor_index.sql`).
- Menginisialisasi permissions RBAC in-memory (*fail-closed*).
- Menjalankan server pada `http://localhost:3000`.

---

## Matriks Endpoint & Perilaku Negosiasi

### 1. Autentikasi (`/api/v1/auth`)
| Metode | Endpoint | Guard | Keterangan |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | Publik | Registrasi pengguna baru dengan validasi username unik & password kuat |
| `POST` | `/api/v1/auth/login` | Publik | Autentikasi username & password, menerbitkan Access Token & Refresh Token |
| `POST` | `/api/v1/auth/refresh` | Publik | Rotasi token dengan Refresh Token yang sah |

### 2. Pengguna (`/api/v1/users`)
| Metode | Endpoint | Guard / Hak Akses | Keterangan |
|---|---|---|---|
| `GET` | `/api/v1/users?limit=5&cursor=...` | `RequireAuth` + `user:list` | Keyset pagination data pengguna (Accept: JSON / CSV) |
| `GET` | `/api/v1/users/:id` | `RequireAuth` + Ownership / `user:read:any` | Detail akun pengguna |
| `POST` | `/api/v1/users` | `RequireAuth` + `user:create` | Pembuatan akun dengan validasi deklaratif |
| `PATCH` | `/api/v1/users/:id` | `RequireAuth` + Ownership / `user:update:any` | Mutasi parsial pengguna dengan pointer `omitnil` |
| `PATCH` | `/api/v1/users/:id/role` | `RequireAuth` + `user:assign_role` | Penetapan role baru (admin tidak boleh ubah role sendiri) |
| `DELETE`| `/api/v1/users/:id` | `RequireAuth` + `user:delete` | Penghapusan akun pengguna |

### 3. Mahasiswa (`/api/v1/students`)
| Metode | Endpoint | Guard / Hak Akses | Keterangan |
|---|---|---|---|
| `GET` | `/api/v1/students?limit=5&cursor=...` | `RequireAuth` + `student:list` | Keyset pagination mahasiswa (Accept: JSON / CSV) |
| `POST` | `/api/v1/students` | `RequireAuth` + `student:create` | Pembuatan mahasiswa (NIM tepat 9 digit, status 201) |
| `GET` | `/api/v1/students/:id` | `RequireAuth` + Ownership / `student:read:any` | Detail mahasiswa |
| `PUT` | `/api/v1/students/:id` | `RequireAuth` + Ownership / `student:update:any` | Penggantian penuh (*full replacement*) data mahasiswa |
| `PATCH` | `/api/v1/students/:id` | `RequireAuth` + Ownership / `student:update:any` | Mutasi parsial mahasiswa via tag `omitnil` |
| `DELETE`| `/api/v1/students/:id` | `RequireAuth` + `student:delete` | Penghapusan data mahasiswa |

---

## Eksekusi Pengujian Unit (Unit Tests)

Jalankan pengujian unit tanpa memerlukan ketergantungan database aktif:

```bash
# Dari dalam direktori Modul_7/api-students:
docker run --rm -v "$PWD":/app -w /app golang:alpine go test -v ./app/service/...

# Dari direktori root workspace Go:
docker run --rm -v "$PWD/Modul_7/api-students":/app -w /app golang:alpine go test -v ./app/service/...
```

Seluruh 11 suite unit test (mencakup registrasi, otorisasi RBAC, kepemilikan data, validasi deklaratif custom NIM, tag `omitnil`, dan mutasi parsial) akan menghasilkan status **PASS**.

---

## Koleksi Postman

Dokumentasi endpoint dan skrip pengujian otomatis tersedia di berkas:
[`api-students-modul7.postman_collection.json`](api-students-modul7.postman_collection.json)

**Fitur Koleksi:**
- Variabel koleksi otomatis: `base_url`, `admin_token`, `staff_token`, `user_token`, `created_student_id`, dan `student_cursor`.
- Test Scripts pada endpoint login dan pagination yang otomatis menyimpan token dan kursor ke variabel koleksi.
- Terbagi dalam 5 folder terstruktur:
  1. `01. Auth & Token Generation`
  2. `02. Declarative Validation (422 Unprocessable Entity)`
  3. `03. Cursor-Based Pagination`
  4. `04. Content Negotiation`
  5. `05. Standardized Error Responses`
