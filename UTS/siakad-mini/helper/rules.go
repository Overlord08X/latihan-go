package helper

import (
	"regexp"
	"time"
)

var (
	nimRegex           = regexp.MustCompile(`^\d{12}$`)
	tahunAkademikRegex = regexp.MustCompile(`^\d{4}/\d{4}-(Ganjil|Genap)$`)
)

// GetMaxSKS menghitung batas maksimal SKS berdasarkan IPK terakhir mahasiswa
// Business rule:
// a. IPK >= 3.00 maksimal 24 SKS
// b. IPK 2.50-2.99 maksimal 21 SKS
// c. IPK < 2.50 maksimal 18 SKS
func GetMaxSKS(ipk float64) int {
	if ipk >= 3.00 {
		return 24
	}
	if ipk >= 2.50 {
		return 21
	}
	return 18
}

// ValidateNIM memeriksa apakah NIM tepat 12 digit numerik
func ValidateNIM(nim string) bool {
	return nimRegex.MatchString(nim)
}

// ValidateAngkatan memeriksa angkatan 4 digit dan <= tahun berjalan
func ValidateAngkatan(angkatan int, currentYear int) bool {
	if currentYear <= 0 {
		currentYear = time.Now().Year()
	}
	return angkatan >= 2000 && angkatan <= currentYear
}

// ValidateTahunAkademik memeriksa format tahun akademik, misal: 2026/2027-Ganjil
func ValidateTahunAkademik(ta string) bool {
	return tahunAkademikRegex.MatchString(ta)
}

// CanAccessStudent menentukan hak akses ke data detail mahasiswa
func CanAccessStudent(role string, loggedInStudentID, targetStudentID int64) bool {
	if role == "admin" {
		return true
	}
	if role == "mahasiswa" && loggedInStudentID > 0 && loggedInStudentID == targetStudentID {
		return true
	}
	return false
}
