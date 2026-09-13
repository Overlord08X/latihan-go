// Package model mendefinisikan entitas domain aplikasi.
package model

import "time"

// Student adalah representasi satu baris pada tabel students di PostgreSQL.
// Tag `db` digunakan oleh pgx untuk memetakan kolom ke field struct.
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// GetInfo mengembalikan ringkasan data mahasiswa (value receiver).
func (s Student) GetInfo() string {
	status := "nonaktif"
	if s.IsActive {
		status = "aktif"
	}
	return s.NIM + " — " + s.Name + " (" + status + ")"
}
