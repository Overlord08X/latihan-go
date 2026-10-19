# api-students  -  Modul 9: Transactions & Concurrency

Backend REST API dengan implementasi **Transactions & Concurrency Control** berstandar industri menggunakan **Go (Fiber v2) + PostgreSQL 16**.  
Proyek ini mengadopsi interface `DBTX`, manajemen transaksi terpusat via `TxManager` dengan retry loop otomatis untuk *serialization failure* (40001) dan *deadlock* (40P01), tiga teknik pengendalian konkurensi (UPDATE bersyarat atomik, pessimistic locking `FOR UPDATE`, dan optimistic locking dengan kolom `version`), pencegahan deadlock melalui penguncian terurut seragam, serta verifikasi konkurensi serentak menggunakan goroutine load test tool.

---

## Fitur Utama Modul 9

1. **Abstraksi Repository Transaksional via Interface `DBTX`:**
   - Menyediakan interface `DBTX` (`Exec`, `Query`, `QueryRow`) yang dipenuhi oleh `*pgxpool.Pool` dan `pgx.Tx`.
   - Menggunakan method `repo.WithTx(tx)` untuk memproduksi salinan repository yang berjalan di dalam konteks transaksi tanpa merusak Clean Architecture.

2. **Manajemen Transaksi Terpusat (`TxManager`):**
   - Satu pintu pembukaan (`pool.Begin`) dan penutupan transaksi dengan idiom `defer tx.Rollback(ctx)`.
   - Mekanisme retry otomatis ber-exponential backoff acak saat terjadi kegagalan retryable seperti deadlock (SQLSTATE `40P01`) atau kegagalan serialisasi (SQLSTATE `40001`).

3. **Tiga Teknik Pengendalian Konkurensi:**
   - **UPDATE Bersyarat Atomik:** Digunakan pada kuota kelas dan kuota pembimbing (`WHERE id = $1 AND enrolled_count < quota`), mengeliminasi race condition *check-then-act* dan *lost update*.
   - **Pessimistic Locking (`SELECT ... FOR UPDATE`):** Mengunci baris `users` mahasiswa untuk menjamin batas 24 SKS tetap konsisten saat request pendaftaran dikirim secara serentak, serta mengunci baris `students` saat penetapan pembimbing.
   - **Optimistic Locking (Kolom `version`):** Diterapkan pada update kuota oleh admin (`WHERE id = $2 AND version = $3`), mencegah modifikasi berbasis data usang (*stale read*) tanpa menahan kunci transaksi.

4. **Pencegahan Deadlock (Deadlock Prevention):**
   - Mengunci resource jamak (seperti pada endpoint swap kelas atau switch pembimbing) selalu dalam urutan ID menaik (`ORDER BY id FOR UPDATE` via clause `ANY($1)`).
   - Menghilangkan siklus saling tunggu sehingga terbebas dari deadlock.

5. **Tugas Mandiri D.2  -  Penetapan & Pergantian Dosen Pembimbing TA:**
   - Pengelolaan entitas `supervisors` dan `supervisions`.
   - Constraint database `PRIMARY KEY (student_id)` menjamin 1 mahasiswa hanya memiliki 1 dosen pembimbing.
   - Endpoint `POST /api/v1/supervisors/:id/assign` dengan UPDATE bersyarat kuota dan kunci row mahasiswa.
   - Endpoint `POST /api/v1/supervisions/switch` atomik dan bebas deadlock.
   - Endpoint `PATCH /api/v1/supervisors/:id` dengan optimistic locking.

---

## Daftar Endpoint Baru

| Method | Endpoint | Hak Akses | Deskripsi |
|---|---|---|---|
| `GET` | `/api/v1/courses` | Authenticated | Mengambil daftar seluruh kelas |
| `GET` | `/api/v1/courses/mine` | Authenticated | Mengambil KRS mahasiswa login dan total SKS |
| `GET` | `/api/v1/courses/:id` | Authenticated | Mengambil detail kelas |
| `POST` | `/api/v1/courses/:id/enroll` | Authenticated | Mendaftar kelas (atomic UPDATE bersyarat + LockStudent) |
| `DELETE` | `/api/v1/courses/:id/enroll` | Authenticated | Membatalkan kelas (Drop) |
| `POST` | `/api/v1/enrollments/swap` | Authenticated | Pindah kelas (atomic + deadlock-free via LockCourses) |
| `PATCH` | `/api/v1/courses/:id` | `course:manage` | Mengubah kuota kelas (Optimistic locking) |
| `GET` | `/api/v1/supervisors` | Authenticated | Mengambil daftar dosen pembimbing |
| `GET` | `/api/v1/supervisors/:id` | Authenticated | Mengambil detail dosen pembimbing |
| `POST` | `/api/v1/supervisors/:id/assign` | `supervision:manage` | Menetapkan pembimbing TA |
| `POST` | `/api/v1/supervisions/switch` | `supervision:manage` | Mengganti pembimbing TA (atomic & deadlock-free) |
| `PATCH` | `/api/v1/supervisors/:id` | `supervision:manage` | Mengubah kuota pembimbing (Optimistic locking) |

---

## Cara Menjalankan dengan Docker

```bash
# 1. Masuk ke direktori
cd Modul_9/api-students

# 2. Jalankan PostgreSQL dan aplikasi melalui Docker Compose
docker compose up -d

# 3. Uji beban menggunakan loadtest tool
docker run --rm --net=host -v "$(pwd)":/app -w /app alpine:latest ./loadtest -mode crowd
docker run --rm --net=host -v "$(pwd)":/app -w /app alpine:latest ./loadtest -mode credits
docker run --rm --net=host -v "$(pwd)":/app -w /app alpine:latest ./loadtest -mode swap
docker run --rm --net=host -v "$(pwd)":/app -w /app alpine:latest ./loadtest -mode supervisor
```
