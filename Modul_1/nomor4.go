package main

import "fmt"

// Student menyimpan data mahasiswa.
type Student struct {
	ID       int
	Name     string
	Grade    float64
	IsActive bool
}

// GetInfo menggunakan value receiver karena method
// hanya membaca data Student dan tidak mengubahnya.
func (s Student) GetInfo() string {
	return fmt.Sprintf(
		"ID: %d, Name: %s, Grade: %.2f, Active: %t",
		s.ID,
		s.Name,
		s.Grade,
		s.IsActive,
	)
}

// UpdateGrade menggunakan pointer receiver karena
// method ini mengubah nilai Grade pada Student.
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate menggunakan pointer receiver karena
// method ini mengubah nilai IsActive.
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate menggunakan pointer receiver karena
// method ini mengubah nilai IsActive.
func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {
	// Membuat objek Student
	student := Student{
		ID:       1,
		Name:     "Raihan",
		Grade:    85.5,
		IsActive: false,
	}

	fmt.Println("=== DATA AWAL STUDENT ===")
	fmt.Println(student.GetInfo())

	// Mengaktifkan student
	student.Activate()

	fmt.Println("\n=== SETELAH ACTIVATE ===")
	fmt.Println(student.GetInfo())

	// Mengubah nilai
	student.UpdateGrade(92.5)

	fmt.Println("\n=== SETELAH UPDATE GRADE ===")
	fmt.Println(student.GetInfo())

	// Menonaktifkan student
	student.Deactivate()

	fmt.Println("\n=== SETELAH DEACTIVATE ===")
	fmt.Println(student.GetInfo())
}