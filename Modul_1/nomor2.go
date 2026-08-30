package main

import "fmt"

func main() {
	// 1. Lima variabel dengan tipe berbeda
	var nama string = "Raihan"
	var umur int = 25
	var ipk float64 = 3.75
	var aktif bool = true
	var mataKuliah []string = []string{
		"Pemrograman Backend Lanjut",
		"Keamanan Cyber",
	}

	fmt.Println("=== DATA VARIABEL ===")
	fmt.Println("Nama:", nama)
	fmt.Println("Umur:", umur)
	fmt.Println("IPK:", ipk)
	fmt.Println("Aktif:", aktif)
	fmt.Println("Mata Kuliah:", mataKuliah)

	// 2. Map untuk menyimpan data mahasiswa
	// Nama sebagai key dan nilai sebagai value
	nilaiMahasiswa := make(map[string]int)

	// 3. Menambahkan data ke map
	nilaiMahasiswa["Raihan"] = 90
	nilaiMahasiswa["Budi"] = 85
	nilaiMahasiswa["Siti"] = 95

	fmt.Println("\n=== DATA MAHASISWA ===")
	fmt.Println(nilaiMahasiswa)

	// 4. Membaca data dengan pengecekan keberadaan
	nilai, ada := nilaiMahasiswa["Raihan"]

	if ada {
		fmt.Println("Nilai Raihan:", nilai)
	} else {
		fmt.Println("Data Raihan tidak ditemukan")
	}

	// Mengecek mahasiswa yang tidak ada
	nilai, ada = nilaiMahasiswa["Andi"]

	if ada {
		fmt.Println("Nilai Andi:", nilai)
	} else {
		fmt.Println("Data Andi tidak ditemukan")
	}

	// 5. Menghapus data dari map
	delete(nilaiMahasiswa, "Budi")

	fmt.Println("\n=== SETELAH BUDI DIHAPUS ===")
	fmt.Println(nilaiMahasiswa)

	// 6. Menelusuri seluruh isi map
	fmt.Println("\n=== SELURUH DATA MAHASISWA ===")

	for nama, nilai := range nilaiMahasiswa {
		fmt.Printf("Nama: %s, Nilai: %d\n", nama, nilai)
	}
}
