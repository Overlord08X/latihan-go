# Praktikum Pemrograman Backend Lanjut (PBE)

Repositori pengerjaan tugas praktikum dan Ujian Tengah Semester (UTS) mata kuliah **Praktikum Pemrograman Backend Lanjut**, Program Studi D4 Teknik Informatika, Fakultas Vokasi, Universitas Airlangga.

- **Nama**: Raihan Zulfa Kamal
- **NIM**: 434241096
- **Program Studi**: D4 Teknik Informatika
- **Tahun Akademik**: 2026/2027

---

## Daftar Modul Praktikum dan Proyek

| No | Modul / Proyek | Topik Utama | Lokasi Direktori |
|:---|:---|:---|:---|
| 1 | **Modul 1** | Persiapan Lingkungan & Sintaks Dasar Go Fiber | [`Modul_1/`](file:///home/raihanzk/Documents/kodingan/Go/Modul_1/) |
| 2 | **Modul 2** | REST API & HTTP Deep Dive | [`Modul_2/api-students/`](file:///home/raihanzk/Documents/kodingan/Go/Modul_2/api-students/) |
| 3 | **Modul 3** | Database Relasional & Repository Pattern | [`Modul_3/api-students/`](file:///home/raihanzk/Documents/kodingan/Go/Modul_3/api-students/) |
| 4 | **Modul 4** | Clean Architecture (Handler, Service, Repository) | [`Modul_4/api-students/`](file:///home/raihanzk/Documents/kodingan/Go/Modul_4/api-students/) |
| 5 | **Modul 5** | Authentication, JWT, & Security Hardening | [`Modul_5/api-students/`](file:///home/raihanzk/Documents/kodingan/Go/Modul_5/api-students/) |
| 6 | **Modul 6** | Authorization & Role-Based Access Control (RBAC) | [`Modul_6/api-students/`](file:///home/raihanzk/Documents/kodingan/Go/Modul_6/api-students/) |
| 7 | **Modul 7** | Advanced API Design (Validation, Keyset Pagination, Content Negotiation) | [`Modul_7/api-students/`](file:///home/raihanzk/Documents/kodingan/Go/Modul_7/api-students/) |
| 8 | **UTS** | **SIAKAD Mini RESTful API (10 Endpoints, SKS Logic, Race Condition Guard)** | [`UTS/siakad-mini/`](file:///home/raihanzk/Documents/kodingan/Go/UTS/siakad-mini/) |

---

## Ujian Tengah Semester (UTS): SIAKAD Mini API

Proyek UTS mengimplementasikan sistem akademik mini (SIAKAD Mini) dengan 10 endpoint lengkap sesuai spesifikasi, dilengkapi dengan:
- **Role-Based Access Control**: `admin` dan `mahasiswa`.
- **Database Row Locking (`SELECT ... FOR UPDATE`)**: Mencegah *race condition* dan *overbooking* kuota mata kuliah saat KRS.
- **Logika Batas SKS Dinamis**: Berbasis IPK (IPK >= 3.00 max 24 SKS, IPK 2.50-2.99 max 21 SKS, IPK < 2.50 max 18 SKS).
- **Soft Delete**: Mahasiswa yang dihapus ditandai dengan `deleted_at` dan diblokir dari login.
- **Rate Limiting**: Maksimal 5 kali kegagalan login per menit mengembalikan status HTTP `429 Too Many Requests`.
- **Containerized Environment**: Berjalan mandiri menggunakan Docker Compose (`siakad_app` & `siakad_db`).

Petunjuk lengkap pengujian dan eksekusi dapat dilihat pada dokumentasi [UTS/siakad-mini/README.md](file:///home/raihanzk/Documents/kodingan/Go/UTS/siakad-mini/README.md).
