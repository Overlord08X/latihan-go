# Praktikum Pemrograman Backend Lanjut

Repository ini berisi hasil praktikum **Pemrograman Backend Lanjut** menggunakan bahasa pemrograman **Go**.

## Environment

Tools yang digunakan:

* Go `1.26.5`
* Git `2.55.0.windows.2`
* PostgreSQL `18.4`
* MongoDB `8.2.12`
* Visual Studio Code
* Go Extension for VS Code
* Fiber v2

## Struktur Project

```text
latihan-go/
├── go.mod
├── go.sum
├── nomor1.go
├── nomor2.go
├── nomor3.go
├── nomor4.go
├── README.md
└── AI-USAGE.md
```

## Daftar Tugas

### Nomor 1 — Persiapan Environment

Melakukan instalasi dan verifikasi environment yang diperlukan untuk praktikum.

Perintah verifikasi:

```powershell
go version
git --version
psql --version
mongod --version
```

Hasil:

```text
go version go1.26.5 windows/amd64
git version 2.55.0.windows.2
psql (PostgreSQL) 18.4
db version v8.2.12
```

### Nomor 2 — Variabel dan Struktur Data

Program mendemonstrasikan:

* Variabel `string`
* Variabel `int`
* Variabel `float64`
* Variabel `bool`
* Slice
* Map data mahasiswa
* Menambahkan data ke map
* Membaca data dengan pengecekan keberadaan
* Menghapus data dari map
* Menelusuri seluruh isi map

File:

```text
nomor2.go
```

### Nomor 3 — Pointer

Program mendemonstrasikan:

* `swap(a, b *int)`
* `updateSlice(s *[]string, newItem string)`
* Pass by value
* Pass by pointer
* Perbedaan perubahan nilai melalui pointer

File:

```text
nomor3.go
```

### Nomor 4 — Struct Student

Program membuat struct:

```go
type Student struct {
    ID       int
    Name     string
    Grade    float64
    IsActive bool
}
```

Method yang digunakan:

* `GetInfo() string`
* `UpdateGrade(grade float64)`
* `Activate()`
* `Deactivate()`

`GetInfo()` menggunakan **value receiver**, sedangkan `UpdateGrade()`, `Activate()`, dan `Deactivate()` menggunakan **pointer receiver** karena ketiga method tersebut mengubah data pada struct.

File:

```text
nomor4.go
```

## Cara Menjalankan

Pastikan Go sudah terpasang, kemudian jalankan masing-masing tugas:

```powershell
go run nomor1.go
go run nomor2.go
go run nomor3.go
go run nomor4.go
```

## Penggunaan AI

Bantuan AI digunakan sebagai pendamping pembelajaran, terutama untuk:

* Memahami instruksi tugas.
* Memahami konsep variabel dan struktur data Go.
* Memahami pointer, pass by value, dan pass by pointer.
* Memahami value receiver dan pointer receiver.
* Membantu memeriksa dan menjelaskan kode.
* Membantu menjelaskan error ketika menjalankan program.

Detail penggunaan AI dicantumkan pada file:

```text
AI-USAGE.md
```

## Repository

GitHub:

https://github.com/Overlord08X/latihan-go
