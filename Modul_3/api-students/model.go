package main

// Request bodies dan response envelope untuk Modul 3.
// Entitas Student sudah dipindahkan ke app/model/student.go.

// -----------------------------------------------------------------------------
// Request bodies
// -----------------------------------------------------------------------------

// CreateRequest adalah body yang diharapkan untuk POST /students.
type CreateRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive *bool   `json:"is_active"`
}

// ReplaceRequest adalah body untuk PUT /students/:id.
// Semua field wajib hadir.
type ReplaceRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}

// PatchRequest adalah body untuk PATCH /students/:id.
// Hanya field yang dikirim yang diubah.
type PatchRequest struct {
	NIM      *string  `json:"nim"`
	Name     *string  `json:"name"`
	Grade    *float64 `json:"grade"`
	IsActive *bool    `json:"is_active"`
}

// -----------------------------------------------------------------------------
// Response envelope
// -----------------------------------------------------------------------------

// Envelope adalah amplop respons seragam untuk seluruh endpoint.
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
