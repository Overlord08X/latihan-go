package main

// Student adalah entitas utama yang mewakili satu mahasiswa.
// Field NIM ditambahkan sebagai penanda unik mahasiswa.
type Student struct {
	ID       int     `json:"id"`
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// GetInfo mengembalikan ringkasan data mahasiswa (value receiver,
// tidak mengubah data).
func (s Student) GetInfo() string {
	status := "nonaktif"
	if s.IsActive {
		status = "aktif"
	}
	return s.NIM + " — " + s.Name + " (grade: " + floatToStr(s.Grade) + ", " + status + ")"
}

// -----------------------------------------------------------------------------
// Request bodies
// -----------------------------------------------------------------------------

// CreateRequest adalah body yang diharapkan untuk POST /students.
// Seluruh field wajib diisi.
type CreateRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive *bool   `json:"is_active"`
}

// ReplaceRequest adalah body untuk PUT /students/:id.
// Semua field wajib hadir; jika ada yang hilang, server menolak dengan 422.
type ReplaceRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}

// PatchRequest adalah body untuk PATCH /students/:id.
// Hanya field yang dikirim yang akan diubah.
type PatchRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}

// -----------------------------------------------------------------------------
// Response envelope
// -----------------------------------------------------------------------------

// Envelope adalah amplop respons seragam untuk semua endpoint.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Meta menyimpan informasi paginasi untuk endpoint daftar.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
