// Package model mendefinisikan struct request dan response untuk API mahasiswa.
// Package ini tidak boleh mengimpor Fiber, repository, atau package lain dari proyek ini.
package model

// CreateStudentRequest adalah body yang diharapkan untuk POST /students.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive *bool   `json:"is_active"`
}

// ReplaceStudentRequest adalah body untuk PUT /students/:id.
// Semua field wajib hadir.
type ReplaceStudentRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}

// PatchStudentRequest adalah body untuk PATCH /students/:id.
// Hanya field yang dikirim yang diubah.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}
