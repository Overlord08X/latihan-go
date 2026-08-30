package main

import "fmt"

// swap menukar dua nilai integer melalui pointer.
func swap(a, b *int) {
	*a, *b = *b, *a
}

// updateSlice menambahkan item baru ke slice melalui pointer.
func updateSlice(s *[]string, newItem string) {
	*s = append(*s, newItem)
}

// passByValue menerima salinan nilai.
func passByValue(x int) {
	x = 100
}

// passByPointer menerima alamat dari sebuah nilai.
func passByPointer(x *int) {
	*x = 100
}

func main() {
	// ========================================
	// 1. Membuktikan function swap
	// ========================================
	fmt.Println("=== SWAP DENGAN POINTER ===")

	a := 10
	b := 20

	fmt.Println("Sebelum swap:")
	fmt.Println("a =", a)
	fmt.Println("b =", b)

	swap(&a, &b)

	fmt.Println("Setelah swap:")
	fmt.Println("a =", a)
	fmt.Println("b =", b)

	// ========================================
	// 2. Membuktikan updateSlice
	// ========================================
	fmt.Println("\n=== UPDATE SLICE DENGAN POINTER ===")

	buah := []string{"Apel", "Jeruk"}

	fmt.Println("Sebelum update:", buah)

	updateSlice(&buah, "Mangga")

	fmt.Println("Setelah update:", buah)

	// ========================================
	// 3. Pass by Value
	// ========================================
	fmt.Println("\n=== PASS BY VALUE ===")

	nilai1 := 50

	fmt.Println("Sebelum function:", nilai1)

	passByValue(nilai1)

	fmt.Println("Setelah function:", nilai1)

	// ========================================
	// 4. Pass by Pointer
	// ========================================
	fmt.Println("\n=== PASS BY POINTER ===")

	nilai2 := 50

	fmt.Println("Sebelum function:", nilai2)

	passByPointer(&nilai2)

	fmt.Println("Setelah function:", nilai2)
}
