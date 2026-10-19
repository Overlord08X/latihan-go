# api-students - Modul 10: API dengan MongoDB (Polyglot Persistence)

Backend REST API dengan implementasi arsitektur **Polyglot Persistence** menggabungkan kekuatan relational database **PostgreSQL 16** dan document-oriented database **MongoDB 7** menggunakan **Go (Fiber v2)** dan official driver `go.mongodb.org/mongo-driver/v2`.

Proyek ini mendemonstrasikan pemisahan tanggung jawab penyimpanan data:
- **PostgreSQL (ACID & Relational):** Menyimpan entitas inti akademik seperti data mahasiswa, otentikasi akun pengguna, role-based access control (RBAC), pendaftaran kelas (KRS), dan pembimbing tugas akhir.
- **MongoDB (Schema-Flexible & Document):** Menyimpan data bersarang (nested/embedded) yang dinamis, tidak memerlukan transaksi lintas tabel yang kaku, serta membutuhkan performa write dan agregasi tinggi: ulasan mata kuliah (*course reviews*) dan portofolio profesional mahasiswa (*student portfolios*).

---

## Fitur Utama Modul 10

1. **Polyglot Persistence & Dual Database Health Check:**
   - Inisialisasi koneksi paralel PostgreSQL (`*pgxpool.Pool`) dan MongoDB (`*mongo.Client`).
   - Endpoint `GET /api/v1/health` melakukan ping ke kedua database secara bersamaan. Mengembalikan status `503 Service Unavailable` secara spesifik jika salah satu database mengalami gangguan jaringan atau down.
   - Penutupan koneksi secara tertib (*graceful shutdown*) menggunakan context timeout 10 detik.

2. **Course Reviews (Ulasan Mata Kuliah):**
   - **Validasi Lintas Database (Cross-Database Validation):** Memastikan mahasiswa terdaftar aktif pada kelas bersangkutan di PostgreSQL sebelum diizinkan menyimpan ulasan di MongoDB.
   - **Pencegahan Ulasan Ganda (Anti-Duplicate):** Unique compound index `{course_id: 1, user_id: 1}` di MongoDB mencegah mahasiswa memberikan ulasan berulang pada mata kuliah yang sama.
   - **Pencegahan Modifikasi Ilegal (Data Tampering):** Update ulasan menggunakan filter atomik `{_id: id, user_id: current_user_id}` sehingga mahasiswa hanya dapat mengubah ulasannya sendiri.
   - **Upvote / Mark Helpful Atomik:** Menggunakan operator `$addToSet: {helpful_users: user_id}` dan `$inc: {helpful_count: 1}` dalam satu operasi untuk mencegah race condition voting.
   - **Tanggapan Resmi (Official Replies):** Menambahkan respon dari staf atau admin dengan batas maksimal 50 tanggapan per ulasan.
   - **Agregasi Ringkasan Rating:** Pipeline agregasi `$facet` untuk menghitung total ulasan, rata-rata rating, dan distribusi rating bintang 1 hingga 5 secara efisien.

3. **Student Portfolios (Tugas Mandiri C):**
   - **Validasi Keberadaan Mahasiswa:** Memverifikasi ID mahasiswa di database relasional PostgreSQL sebelum membuat atau memperbarui portofolio di MongoDB.
   - **Compound & Multikey Indexing:**
     - Unique index `{student_id: 1}` memastikan satu mahasiswa hanya memiliki satu dokumen portofolio.
     - Multikey index `{projects.technologies: 1}` mempercepat pencarian portofolio berdasarkan keahlian teknologi.
   - **Operasi Atomik Array Proyek:**
     - Tambah proyek (`POST /api/v1/students/:id/portfolio/projects`) dengan batasan kapasitas aman 20 proyek via filter atomik `"projects.19": {"$exists": false}` dan operator `$push`.
     - Hapus proyek (`DELETE /api/v1/students/:id/portfolio/projects/:projectId`) menggunakan operator `$pull` berdasarkan ID proyek tanpa perlu membaca dan menulis ulang seluruh dokumen.
   - **Agregasi Statistik Teknologi Terpopuler:** Pipeline agregasi dengan tahapan ganda `$unwind` (`$projects` kemudian `$projects.technologies`), `$group`, `$sort`, dan `$limit` untuk mengambil 10 keahlian teknologi paling populer di kalangan mahasiswa.

---

## Daftar Endpoint Modul 10

| Method | Endpoint | Hak Akses | Deskripsi |
|---|---|---|---|
| `GET` | `/api/v1/health` | Publik | Memeriksa ketersediaan server, PostgreSQL, dan MongoDB |
| `POST` | `/api/v1/auth/login` | Publik | Otentikasi pengguna (JWT token) |
| `POST` | `/api/v1/courses/:id/reviews` | Authenticated | Menulis ulasan mata kuliah (harus terdaftar di PG) |
| `GET` | `/api/v1/courses/:id/reviews` | Authenticated | Mengambil daftar ulasan (pagination & filter tag) |
| `GET` | `/api/v1/courses/:id/reviews/summary` | Authenticated | Mengambil ringkasan dan distribusi rating ulasan |
| `PUT` | `/api/v1/reviews/:id` | Authenticated (Owner) | Memperbarui ulasan milik sendiri |
| `POST` | `/api/v1/reviews/:id/helpful` | Authenticated | Memberikan voting bermanfaat (helpful) |
| `POST` | `/api/v1/reviews/:id/replies` | `review:moderate` | Memberikan tanggapan resmi dari staf/admin |
| `DELETE` | `/api/v1/reviews/:id` | `review:moderate` | Menghapus ulasan tidak layak (moderasi admin) |
| `GET` | `/api/v1/students/:id/portfolio` | Authenticated | Mengambil portofolio lengkap mahasiswa |
| `PUT` | `/api/v1/students/:id/portfolio` | Authenticated | Membuat / memperbarui seluruh portofolio mahasiswa |
| `POST` | `/api/v1/students/:id/portfolio/projects` | Authenticated | Menambahkan satu proyek ke portofolio (maksimal 20) |
| `DELETE` | `/api/v1/students/:id/portfolio/projects/:projectId` | Authenticated | Menghapus proyek dari portofolio |
| `GET` | `/api/v1/portfolios/stats/technologies` | Authenticated | Mengambil statistik 10 teknologi terpopuler |

---

## Menjalankan dengan Docker

Seluruh layanan (PostgreSQL, MongoDB, dan Aplikasi Go) dijalankan di dalam container Docker.

### 1. Menjalankan Container

```bash
# Pastikan berada pada direktori Modul_10/api-students
cd Modul_10/api-students

# Menjalankan database dan aplikasi via Docker Compose
docker compose up -d
```

### 2. Memeriksa Status Container

```bash
docker compose ps
```

Hasil yang diharapkan:
- `praktikum_db`: Port `5432:5432` (healthy)
- `praktikum_mongo`: Port `27017:27017` (running)
- `api_students_m10`: Port `3000:3000` (running)

### 3. Menguji Health Check

```bash
curl -s http://localhost:3000/api/v1/health | python3 -m json.tool
```

Output sukses:
```json
{
  "success": true,
  "message": "server, PostgreSQL, dan MongoDB berjalan"
}
```

---

## Pengujian dengan Postman

File koleksi Postman telah disediakan pada direktori proyek:
- File: `api-students-modul10.postman_collection.json`

### Langkah Import Postman:
1. Buka aplikasi Postman.
2. Klik tombol **Import** pada panel kiri atas.
3. Pilih file `Modul_10/api-students/api-students-modul10.postman_collection.json`.
4. Seluruh folder request (Health Check, Auth, Reviews, Portfolios, dan Negative Cases) siap dijalankan secara langsung.
