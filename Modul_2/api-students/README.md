# api-students — Modul 2: REST API & HTTP Deep Dive

API mahasiswa sederhana menggunakan **Go + Fiber v2** dengan penyimpanan in-memory.

## Environment

- Go `1.26.5`
- Fiber v2 `v2.52.15`
- Port: `3000`

## Cara Menjalankan

```powershell
go run .
```

## Kontrak API

### Struktur Respons Seragam

Semua endpoint mengembalikan amplop JSON berikut:

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

Mengembalikan daftar mahasiswa dengan dukungan paginasi, pencarian, pengurutan, dan filter.

**Query string:**

| Parameter  | Default | Keterangan |
|------------|---------|-----------|
| `page`     | 1       | Nomor halaman |
| `limit`    | 10      | Jumlah per halaman (maks. 100) |
| `search`   | —       | Cari berdasarkan nama (case-insensitive) |
| `sort`     | id      | Urut berdasarkan: `id`, `nim`, `name`, `grade` |
| `order`    | asc     | `asc` atau `desc` |
| `is_active`| —       | Filter: `true` atau `false` |

**Contoh request:**

```
GET /api/v1/students?page=1&limit=2&sort=name&order=asc&is_active=true
```

**Status:** `200`

**Contoh respons:**

```json
{
  "success": true,
  "message": "daftar mahasiswa",
  "data": [
    { "id": 1, "nim": "2023001", "name": "Andi Saputra", "grade": 85.5, "is_active": true }
  ],
  "meta": { "page": 1, "limit": 2, "total": 2, "total_pages": 1 }
}
```

---

### GET /api/v1/students/:id

Mengembalikan satu mahasiswa berdasarkan ID.

**Contoh request:**

```
GET /api/v1/students/1
```

**Status yang mungkin:** `200`, `400`, `404`

**Contoh respons (200):**

```json
{
  "success": true,
  "message": "data mahasiswa",
  "data": { "id": 1, "nim": "2023001", "name": "Andi Saputra", "grade": 85.5, "is_active": true }
}
```

---

### POST /api/v1/students

Menambahkan mahasiswa baru. Mengembalikan 201 disertai header `Location`.

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

**Contoh respons (201):**

```json
{
  "success": true,
  "message": "data berhasil ditambahkan",
  "data": { "id": 4, "nim": "2023010", "name": "Dian Pratiwi", "grade": 88, "is_active": true }
}
```

Header: `Location: /api/v1/students/4`

---

### PUT /api/v1/students/:id

Mengganti **seluruh** data mahasiswa. Semua field wajib dikirim.

**Header wajib:** `Content-Type: application/json`

**Body:**

```json
{
  "nim": "2023001",
  "name": "Andi Saputra Baru",
  "grade": 90.0,
  "is_active": false
}
```

**Status yang mungkin:** `200`, `400`, `404`, `409`, `415`, `422`

---

### PATCH /api/v1/students/:id

Mengubah **sebagian** data mahasiswa. Hanya field yang dikirim yang berubah.

**Header wajib:** `Content-Type: application/json`

**Body (contoh hanya ubah is_active):**

```json
{
  "is_active": true
}
```

**Status yang mungkin:** `200`, `400`, `404`, `409`, `415`, `422`

---

### DELETE /api/v1/students/:id

Menghapus mahasiswa. Mengembalikan 204 tanpa body.

**Contoh request:**

```
DELETE /api/v1/students/1
```

**Status yang mungkin:** `204`, `400`, `404`

---

## Ringkasan Status HTTP

| Status | Situasi |
|--------|---------|
| 200 | Pengambilan atau perubahan berhasil |
| 201 | Penambahan berhasil, disertai header `Location` |
| 204 | Penghapusan berhasil, tanpa body |
| 400 | Body bukan JSON yang sah, atau id bukan angka |
| 404 | Data tidak ditemukan |
| 409 | NIM sudah digunakan (konflik) |
| 415 | `Content-Type` bukan `application/json` |
| 422 | Validasi isi gagal, dengan rincian per field |

## Penggunaan AI

Bantuan AI digunakan sebagai pendamping pembelajaran untuk:

- Memahami perbedaan PUT (idempoten, ganti seluruh resource) dan PATCH (ganti sebagian).
- Memahami penggunaan pointer pada struct request agar field opsional bisa dibedakan dengan zero value.
- Memahami penggunaan `sync.RWMutex` untuk keamanan akses bersamaan pada store in-memory.
- Memahami kapan menggunakan middleware global vs per-grup.
- Membantu memeriksa kode dan menjelaskan error.
