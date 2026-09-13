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

// CreateNilaiRequest adalah body yang diharapkan untuk POST /nilais.
type CreateNilaiRequest struct {
	IDStudent  *int     `json:"id_student"`
	NamaMatkul string   `json:"nama_matkul"`
	Nilai      *float64 `json:"nilai"`
}

// RegisterRequest adalah body untuk POST /auth/register.
// Sengaja TIDAK memuat field role — role selalu ditentukan server ("user").
// Ini menutup kerentanan mass assignment.
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest adalah body untuk POST /auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshRequest adalah body untuk POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest adalah body untuk POST /auth/logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
