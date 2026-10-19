package model

import "time"

// Nilai mempresentasikan data nilai mata kuliah mahasiswa di database
type Nilai struct {
	IDNilai    int       `json:"id_nilai"`
	IDStudent  int       `json:"id_student"`
	NamaMatkul string    `json:"nama_matkul"`
	Nilai      float64   `json:"nilai"`
	CreatedAt  time.Time `json:"created_at"`
}