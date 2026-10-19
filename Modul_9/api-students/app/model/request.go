// Package model mendefinisikan struct request dan response untuk API mahasiswa.
// Package ini tidak boleh mengimpor Fiber, repository, atau package lain dari proyek ini.
package model

// CreateStudentRequest adalah body deklaratif untuk POST /students (Tugas Mandiri D.2).
type CreateStudentRequest struct {
	NIM      string  `json:"nim"       validate:"required,nim"`
	Name     string  `json:"name"      validate:"required,min=3,max=100"`
	Grade    float64 `json:"grade"     validate:"min=0,max=100"`
	IsActive *bool   `json:"is_active" validate:"required"`
}

// ReplaceStudentRequest adalah body deklaratif untuk PUT /students/:id (Tugas Mandiri D.2).
// Semua field wajib hadir.
type ReplaceStudentRequest struct {
	NIM      string   `json:"nim"       validate:"required,nim"`
	Name     string   `json:"name"      validate:"required,min=3,max=100"`
	Grade    *float64 `json:"grade"     validate:"required,min=0,max=100"`
	IsActive *bool    `json:"is_active" validate:"required"`
}

// PatchStudentRequest adalah body deklaratif untuk PATCH /students/:id (Tugas Mandiri D.2).
// Menggunakan pointer dan tag omitnil agar field yang tidak dikirim dilewati,
// tetapi jika dikirim bernilai kosong (contoh {"name":""}) tetap diperiksa dan ditolak (422).
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"       validate:"omitnil,nim"`
	Name     *string  `json:"name,omitempty"      validate:"omitnil,min=3,max=100"`
	Grade    *float64 `json:"grade,omitempty"     validate:"omitnil,min=0,max=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// StudentCursorQuery menampung parameter query cursor pagination untuk students (Tugas Mandiri D.3).
type StudentCursorQuery struct {
	Search   string
	IsActive *bool
	After    *Cursor
	Limit    int
}

// CreateNilaiRequest adalah body yang diharapkan untuk POST /nilais.
type CreateNilaiRequest struct {
	IDStudent  *int     `json:"id_student"  validate:"required"`
	NamaMatkul string   `json:"nama_matkul" validate:"required,min=3,max=100"`
	Nilai      *float64 `json:"nilai"       validate:"required,min=0,max=100"`
}

// RegisterRequest adalah body untuk POST /auth/register (Langkah 6).
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=30,username"`
	Email    string `json:"email"    validate:"required,email,max=120"`
	Password string `json:"password" validate:"required,max=72,strongpassword"`
}

// LoginRequest adalah body untuk POST /auth/login.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest adalah body untuk POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest adalah body untuk POST /auth/logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
