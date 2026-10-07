package service

import (
	"testing"

	"siakad-mini/helper"
)

func TestGetMaxSKS(t *testing.T) {
	tests := []struct {
		name     string
		ipk      float64
		expected int
	}{
		{"IPK 4.00 (Tier S)", 4.00, 24},
		{"IPK 3.85 (Tier A)", 3.85, 24},
		{"IPK 3.00 batas bawah Tier A", 3.00, 24},
		{"IPK 2.99 batas atas Tier B", 2.99, 21},
		{"IPK 2.75 (Tier B)", 2.75, 21},
		{"IPK 2.50 batas bawah Tier B", 2.50, 21},
		{"IPK 2.49 batas atas Tier C", 2.49, 18},
		{"IPK 2.00 (Tier C)", 2.00, 18},
		{"IPK 0.00 (Tier C)", 0.00, 18},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := helper.GetMaxSKS(tc.ipk)
			if got != tc.expected {
				t.Errorf("GetMaxSKS(%.2f) = %d, expected %d", tc.ipk, got, tc.expected)
			}
		})
	}
}

func TestValidateNIM(t *testing.T) {
	tests := []struct {
		name     string
		nim      string
		expected bool
	}{
		{"Valid 12 digit", "434241096001", true},
		{"Valid 12 digit mahasiswa 20", "434241096020", true},
		{"Invalid kurang dari 12 digit (9 digit)", "434241096", false},
		{"Invalid lebih dari 12 digit", "43424109600001", false},
		{"Invalid mengandung huruf", "43424109600A", false},
		{"Invalid mengandung spasi", "434241 096001", false},
		{"Invalid string kosong", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := helper.ValidateNIM(tc.nim)
			if got != tc.expected {
				t.Errorf("ValidateNIM(%s) = %v, expected %v", tc.nim, got, tc.expected)
			}
		})
	}
}

func TestValidateAngkatan(t *testing.T) {
	currentYear := 2026
	tests := []struct {
		name     string
		angkatan int
		expected bool
	}{
		{"Angkatan tahun berjalan (2026)", 2026, true},
		{"Angkatan 2024", 2024, true},
		{"Angkatan 2020", 2020, true},
		{"Angkatan masa depan (2027)", 2027, false},
		{"Angkatan terlalu lampau (1998)", 1998, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := helper.ValidateAngkatan(tc.angkatan, currentYear)
			if got != tc.expected {
				t.Errorf("ValidateAngkatan(%d, %d) = %v, expected %v", tc.angkatan, currentYear, got, tc.expected)
			}
		})
	}
}

func TestValidateTahunAkademik(t *testing.T) {
	tests := []struct {
		name     string
		ta       string
		expected bool
	}{
		{"Valid Ganjil", "2026/2027-Ganjil", true},
		{"Valid Genap", "2026/2027-Genap", true},
		{"Invalid format tanpa semester", "2026/2027", false},
		{"Invalid semester tidak dikenal", "2026/2027-Pendek", false},
		{"Invalid tahun tunggal", "2026-Ganjil", false},
		{"Invalid kosong", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := helper.ValidateTahunAkademik(tc.ta)
			if got != tc.expected {
				t.Errorf("ValidateTahunAkademik(%s) = %v, expected %v", tc.ta, got, tc.expected)
			}
		})
	}
}

func TestCanAccessStudent(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		authStudentID  int64
		targetStudentID int64
		expected       bool
	}{
		{"Admin akses mahasiswa 1", "admin", 0, 1, true},
		{"Admin akses mahasiswa 2", "admin", 0, 2, true},
		{"Mahasiswa 1 akses dirinya sendiri", "mahasiswa", 1, 1, true},
		{"Mahasiswa 2 akses dirinya sendiri", "mahasiswa", 2, 2, true},
		{"Mahasiswa 1 akses mahasiswa 2 (IDOR ditolak)", "mahasiswa", 1, 2, false},
		{"Mahasiswa 2 akses mahasiswa 1 (IDOR ditolak)", "mahasiswa", 2, 1, false},
		{"Role tidak dikenal", "guest", 1, 1, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := helper.CanAccessStudent(tc.role, tc.authStudentID, tc.targetStudentID)
			if got != tc.expected {
				t.Errorf("CanAccessStudent(%s, %d, %d) = %v, expected %v", tc.role, tc.authStudentID, tc.targetStudentID, got, tc.expected)
			}
		})
	}
}
