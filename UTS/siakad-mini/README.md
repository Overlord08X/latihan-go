# SIAKAD Mini API - Ujian Tengah Semester (UTS) Pemrograman Backend Lanjut

Repositori implementasi RESTful API Back End untuk **SIAKAD Mini**, layanan akademik sederhana yang mengelola data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS). Dibuat dengan bahasa pemrograman **Go (Golang)** menggunakan framework **Fiber v2** dan basis data relasional **PostgreSQL 16**.

---

## 1. Arsitektur dan Fitur Utama

Aplikasi ini dirancang mengikuti prinsip Clean Architecture berlapis (*Controller/Handler*, *Service*, *Repository*, *Model*) dengan fitur utama:
1. **Autentikasi & Otorisasi Berbasis JWT (HS256)**:
   - Token bearer wajib disertakan pada seluruh endpoint privat.
   - Pembedaan hak akses berbasis peran (*Role-Based Access Control*): `admin` dan `mahasiswa`.
   - Proteksi IDOR (*Insecure Direct Object Reference*): Mahasiswa hanya dapat melihat detail diri sendiri dan membatalkan KRS milik sendiri.
2. **Rate Limiting Login**:
   - Membatasi kegagalan login hingga maksimal 5 kali per menit per alamat IP/klien.
   - Mengembalikan HTTP status `429 Too Many Requests` saat batas terlampaui.
3. **Database Concurrency & Row-Level Locking (`SELECT ... FOR UPDATE`)**:
   - Menghindari *race condition* atau *overbooking* kuota mata kuliah saat banyak mahasiswa mengambil KRS secara simultan dalam satu transaksi atomik.
4. **Validasi Deklaratif & Custom Rules**:
   - NIM tepat 12 digit numerik.
   - Angkatan 4 digit dan tidak melebihi tahun berjalan.
   - Format tahun akademik baku: `YYYY/YYYY-(Ganjil|Genap)`.
   - Format IPK 0.00 s.d. 4.00.
5. **Business Rules Batas SKS Berbasis IPK**:
   - IPK >= 3.00: Maksimal 24 SKS.
   - IPK 2.50 - 2.99: Maksimal 21 SKS.
   - IPK < 2.50: Maksimal 18 SKS.
   - Mencegah pengambilan mata kuliah yang sama dua kali pada tahun akademik yang sama (`409 Conflict`).
   - Menyertakan sisa SKS yang dapat diambil dalam pesan kesalahan saat batas terlampaui (`422 Unprocessable Entity`).
6. **Soft Delete Mahasiswa**:
   - Menandai kolom `deleted_at` tanpa menghapus data secara fisik dari basis data.
   - Mahasiswa yang telah di-*soft delete* diisolasi dari daftar mahasiswa dan ditolak saat melakukan login (`401 Unauthorized`).
7. **Penyimpanan Password Aman**:
   - Seluruh kata sandi dienkripsi menggunakan algoritma `bcrypt` dengan cost factor 10.
8. **Format Response JSON Seragam**:
   - Menggunakan amplop standar `{ "success": bool, "message": string, "data": ... }` untuk respon sukses.
   - Menggunakan format `{ "success": false, "message": string, "errors": ... }` untuk respon error validasi (422) maupun error lainnya.
   - Mendukung kode status HTTP standar: `200`, `201`, `204`, `401`, `403`, `404`, `409`, `422`, `429`, dan `500`.

---

## 2. Struktur Direktori Proyek

```text
siakad-mini/
├── app/
│   ├── model/                  # Entitas basis data dan Request DTO
│   │   ├── course.go
│   │   ├── enrollment.go
│   │   ├── request.go
│   │   ├── student.go
│   │   └── user.go
│   ├── repository/             # Lapisan akses basis data SQL (pgxpool)
│   │   ├── course_repository.go
│   │   ├── enrollment_repository.go
│   │   ├── student_repository.go
│   │   └── user_repository.go
│   └── service/                # Lapisan logika bisnis dan orkestrasi
│       ├── auth_service.go
│       ├── course_service.go
│       ├── enrollment_service.go
│       ├── rules_test.go       # Unit test logika bisnis SKS & validasi
│       └── student_service.go
├── config/                     # Konfigurasi aplikasi, env, logger, error handler
│   ├── app.go
│   ├── env.go
│   └── logger.go
├── database/                   # Koneksi pool PostgreSQL & migration runner
│   └── postgres.go
├── helper/                     # Helper respon JSON, AppError, rules, validator
│   ├── errors.go
│   ├── response.go
│   ├── rules.go
│   └── validator.go
├── jwtutil/                    # Utilitas token JWT (Sign & Parse Claims)
│   └── jwt.go
├── middleware/                 # Middleware JWT Auth, Role Guard, Rate Limiter
│   ├── auth.go
│   └── rate_limit.go
├── migrations/                 # Skrip DDL schema dan data seeder awal
│   ├── 001_create_schema.sql
│   └── 002_seed_data.sql
├── route/                      # Registrasi routing Fiber v2
│   ├── auth.go
│   ├── course.go
│   ├── enrollment.go
│   └── student.go
├── .dockerignore
├── .env.example
├── docker-compose.yml          # Konfigurasi container App & PostgreSQL 16
├── Dockerfile                  # Multi-stage build container Go
├── go.mod
├── go.sum
├── main.go                     # Titik masuk utama aplikasi
├── README.md                   # Dokumentasi teknis proyek
└── siakad-mini.postman_collection.json # Koleksi pengujian Postman siap impor
```

---

## 3. Data Awal & Akun Pengujian (Seeder)

Aplikasi telah dilengkapi seeder otomatis saat container pertama kali dijalankan:

### A. Akun Admin
- **Email**: `admin@siakad.ac.id`
- **Password**: `Admin123!`
- **Peran**: `admin`

### B. Akun Mahasiswa Utama (Total 20 Akun)
Password default setiap mahasiswa sama dengan **NIM** masing-masing.

| No | NIM | Nama Mahasiswa | Email | Prodi | Angkatan | IPK | Batas SKS | Password Default |
|:---|:---|:---|:---|:---|:---|:---|:---|:---|
| 1 | `434241096001` | Mahasiswa 01 | `mhs01@siakad.ac.id` | D4 Teknik Informatika | 2024 | 3.85 | 24 | `434241096001` |
| 2 | `434241096002` | Ahmad Fauzi | `mhs02@siakad.ac.id` | D4 Teknik Informatika | 2024 | 3.50 | 24 | `434241096002` |
| 3 | `434241096003` | Budi Santoso | `mhs03@siakad.ac.id` | D4 Teknik Informatika | 2023 | 2.80 | 21 | `434241096003` |
| 4 | `434241096004` | Citra Dewi | `mhs04@siakad.ac.id` | Sistem Informasi | 2024 | 2.30 | 18 | `434241096004` |
| 5 | `434241096005` | Dimas Pratama | `mhs05@siakad.ac.id` | Sistem Informasi | 2023 | 3.75 | 24 | `434241096005` |
| ... | ... | *(s.d. Mahasiswa 20)* | ... | ... | ... | ... | ... | *(NIM)* |

> **Catatan Kasus Uji Khusus pada Data Seeder:**
> 1. **Uji Kuota Penuh (422)**: Mata kuliah `MK009` (*Kapita Selekta Informatika*) dikonfigurasi berkuota 2 dan telah terisi 2 mahasiswa pada semester `2026/2027-Ganjil`.
> 2. **Uji Batas SKS Terlampaui (422)**: Mahasiswa Citra Dewi (`mhs04@siakad.ac.id`, IPK 2.30, batas 18 SKS) telah terdaftar 17 SKS pada semester `2026/2027-Ganjil`. Mengambil mata kuliah 3 SKS tambahan akan memicu penolakan 422 disertai info sisa SKS.
> 3. **Uji Duplikasi KRS (409)**: Mahasiswa 01 (`mhs01@siakad.ac.id`) telah mengambil `MK001` pada semester `2026/2027-Ganjil`. Mengambil kembali akan memicu 409 Conflict.

---

## 4. Cara Menjalankan Aplikasi

### Menjalankan dengan Docker Compose (Direkomendasikan)

Pastikan Docker Engine telah berjalan di komputer Anda, lalu jalankan perintah berikut dari direktori `UTS/siakad-mini`:

```bash
# 1. Jalankan container database PostgreSQL dan API service
docker compose up --build -d

# 2. Periksa status container
docker compose ps

# 3. Pantau log migrasi dan server
docker compose logs -f app
```

API akan aktif dan siap menerima request pada alamat:
```text
http://localhost:3000
```

Untuk menghentikan dan mereset basis data ke kondisi awal seeder:
```bash
docker compose down -v
docker compose up -d
```

### Menjalankan Unit Test

Untuk menjalankan rangkaian unit test logika bisnis (SKS, validasi NIM, validasi tahun akademik):

```bash
docker run --rm -v "$PWD":/app -w /app golang:alpine go test -v ./app/service/...
```

---

## 5. Ringkasan 10 Endpoint API

| No | Method | Endpoint | Akses / Peran | Deskripsi / Fungsi | Kode Respon Sukses |
|:---|:---|:---|:---|:---|:---|
| 1 | `POST` | `/api/v1/auth/login` | Publik | Autentikasi akun, mengembalikan access token JWT (Rate Limit: 5 fails/menit) | `200 OK` |
| 2 | `GET` | `/api/v1/auth/me` | Semua Role | Menampilkan identitas user login (termasuk profil student jika mahasiswa) | `200 OK` |
| 3 | `GET` | `/api/v1/students` | Admin | Menampilkan daftar mahasiswa dengan paginasi, filter prodi/angkatan, pencarian, dan sorting | `200 OK` |
| 4 | `POST` | `/api/v1/students` | Admin | Mendaftarkan mahasiswa baru sekaligus akun user dalam satu transaksi atomik | `201 Created` |
| 5 | `GET` | `/api/v1/students/{id}` | Admin / Mahasiswa (ID Sendiri) | Menampilkan detail mahasiswa, daftar mata kuliah KRS, total SKS, dan batas SKS | `200 OK` |
| 6 | `PUT` | `/api/v1/students/{id}` | Admin | Memperbarui profil mahasiswa (nama, prodi, angkatan, IPK; NIM tidak dapat diubah) | `200 OK` |
| 7 | `DELETE` | `/api/v1/students/{id}` | Admin | Soft delete mahasiswa (mengisi `deleted_at`, isolasi data, dan blokir login) | `204 No Content` |
| 8 | `GET` | `/api/v1/courses` | Semua Role | Menampilkan katalog mata kuliah beserta kalkulasi otomatis `terisi` dan `sisa_kuota` | `200 OK` |
| 9 | `POST` | `/api/v1/enrollments` | Mahasiswa | Mengambil mata kuliah KRS (Pengecekan transaksi dengan row locking kuota & batas SKS) | `201 Created` |
| 10 | `DELETE` | `/api/v1/enrollments/{id}` | Mahasiswa (Milik Sendiri) | Membatalkan pengambilan KRS dan mengembalikan kuota mata kuliah | `204 No Content` |

---

## 6. Petunjuk Import Koleksi Postman

Berkas koleksi Postman yang lengkap dan siap pakai tersedia pada:
```text
UTS/siakad-mini/siakad-mini.postman_collection.json
```

### Langkah Impor:
1. Buka aplikasi **Postman**.
2. Klik tombol **Import** di pojok kiri atas.
3. Seret (*drag-and-drop*) atau pilih berkas `siakad-mini.postman_collection.json`.
4. Koleksi **SIAKAD Mini API - UTS Pemrograman Backend Lanjut** akan muncul dengan 4 folder terstruktur:
   - `01. Authentication & Profiling` (8 request)
   - `02. Manajemen Mahasiswa (Khusus Admin)` (11 request)
   - `03. Akses Mahasiswa & KRS` (5 request)
   - `04. Pengambilan KRS (Enrollments)` (8 request)
5. Seluruh token (`admin_token`, `mhs_token`, `citra_token`) serta ID dinamis (`created_student_id`, `enrolled_id`) otomatis tersimpan ke variabel koleksi Postman saat request dijalankan berurutan.
