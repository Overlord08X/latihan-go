# api-students - Modul 11: Background Jobs & Asynchronous Processing

Backend REST API dan Worker Service dengan implementasi **Transactional Outbox Pattern** dan **Distributed Job Queue** menggunakan **Go (Fiber v2)**, **PostgreSQL 16**, **Redis 7**, dan **Asynq**.

Proyek ini mendemonstrasikan pemisahan tanggung jawab (*separation of concerns*) antara penerimaan permintaan pengguna (kasir / API) dan tugas-tugas komputasi atau I/O pihak ketiga yang berlatensi tinggi (barista / Worker).

---

## Fitur Utama Modul 11

1. **Transactional Outbox Pattern:**
   - Menyimpan rekaman event akademik (pendaftaran KRS, pembatalan kelas, swap kelas, perubahan kuota) secara atomik di tabel `outbox` PostgreSQL dalam satu blok transaksi database yang sama dengan mutasi data bisnis.
   - Mengeliminasi masalah *Dual Write* antara PostgreSQL dan Redis tanpa two-phase commit (2PC).

2. **Outbox Relay Service:**
   - Membaca baris event outbox yang belum dikirim (`published_at IS NULL`) secara berkala dan memindahkannya ke antrean Redis.
   - Menggunakan klausa `FOR UPDATE SKIP LOCKED` untuk mendukung konkurensi multi-worker tanpa lock contention atau duplikasi proses.

3. **Asynchronous Notification Worker:**
   - Memproses pengiriman email notifikasi di latar belakang tanpa menahan lock database utama.
   - Menurunkan waktu respons pendaftaran KRS dari 1.500 ms menjadi rata-rata 3,7 ms.

4. **Idempotensi & Composite Key:**
   - Menggunakan batasan unik komposit `UNIQUE (outbox_id, user_id)` pada tabel `notifications` sebagai idempotency key.
   - Menjamin setiap mahasiswa hanya menerima tepat satu email per event, kebal terhadap duplikasi pengiriman tugas (*at-least-once delivery*).

5. **Klasifikasi Kegagalan & Retry Control:**
   - **Temporary Failure:** Mengaktifkan mekanisme backoff retry otomatis dengan jeda eksponensial (2, 4, 6, 8, 10 detik).
   - **Permanent Failure:** Menggunakan `asynq.SkipRetry` untuk langsung mengarsipkan tugas yang gagal permanen (misal: domain email `.invalid`) tanpa membuang kuota retry.

6. **Broadcast Perubahan Kuota (Tugas Mandiri D.2):**
   - Mengirimkan email notifikasi ke seluruh peserta kelas yang terdaftar saat admin mengubah kuota mata kuliah (`PATCH /api/v1/courses/:id`).
   - Setiap penerima dikelola dalam transaksi independen (`WithTxOnce`), menjamin bahwa kegagalan pengiriman pada satu mahasiswa tidak membatalkan atau mengirim ulang email ke mahasiswa lainnya.

7. **Single Binary Multi-Role CLI & Graceful Shutdown:**
   - Satu kode sumber yang dapat dijalankan sebagai API (`-role=api`), worker (`-role=worker`), atau keduanya (`-role=all`).
   - Menangkap sinyal OS (`SIGINT`, `SIGTERM`) untuk menyelesaikan pengiriman email yang sedang berjalan secara utuh sebelum proses dihentikan.

---

## Daftar Endpoint Terkait Notifikasi & Monitoring

| Method | Endpoint | Hak Akses | Deskripsi |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/notifications/mine` | Authenticated | Mengambil riwayat notifikasi milik mahasiswa yang sedang login |
| `GET` | `/api/v1/admin/queue` | `queue:monitor` | Memantau statistik antrean outbox PostgreSQL dan queue Redis Asynq |
| `POST` | `/api/v1/courses/:id/enroll` | Authenticated | Pendaftaran kelas (menghasilkan event `enrollment.created`) |
| `DELETE` | `/api/v1/courses/:id/enroll` | Authenticated | Pembatalan kelas (menghasilkan event `enrollment.dropped`) |
| `POST` | `/api/v1/enrollments/swap` | Authenticated | Pindah kelas (menghasilkan event `enrollment.swapped`) |
| `PATCH` | `/api/v1/courses/:id` | `course:manage` | Mengubah kuota kelas (menghasilkan event `course.quota_changed`) |

---

## Panduan Menjalankan Layanan

### 1. Menjalankan Infrastruktur Basis Data
Pastikan PostgreSQL, MongoDB, dan Redis telah berjalan (dapat menggunakan Docker):
```bash
docker start praktikum_db praktikum_mongo praktikum_redis
```

### 2. Menjalankan Proses API dan Worker
Buka dua sesi terminal di folder `Modul_11/api-students`:
```bash
# Terminal 1: Menjalankan Server HTTP API
go run . -role=api

# Terminal 2: Menjalankan Background Worker & Outbox Relay
go run . -role=worker
```

Atau jalankan keduanya dalam satu proses terpadu untuk pengujian lokal:
```bash
go run . -role=all
```

---

## Pengujian Idempotensi dengan Alat Redeliver
Gunakan alat bantu `tools/redeliver` untuk mensimulasikan duplikasi task pada antrean Redis:
```bash
go run ./tools/redeliver -outbox <ID_OUTBOX> -copies 5
```
