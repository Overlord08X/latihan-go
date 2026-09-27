package service

import (
	"strings"

	"api-students/app/model"
)

// ApplyPatch menerapkan perubahan parsial dari PatchUserRequest ke User yang ada (Langkah 6).
// Pemeriksaan bentuk sudah selesai dikerjakan tag validator sebelum fungsi ini dipanggil,
// sehingga di sini tugasnya tinggal satu: menggabungkan nilai yang dikirim.
func ApplyPatch(current model.User, req model.PatchUserRequest) model.User {
	if req.Username != nil {
		current.Username = strings.TrimSpace(*req.Username)
	}
	if req.Email != nil {
		current.Email = strings.TrimSpace(*req.Email)
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

// IsEmptyPatch memeriksa body PATCH yang tidak berisi field apa pun (Langkah 6).
// Aturan ini tidak dapat ditulis sebagai tag: tag memeriksa satu field
// pada satu waktu, sedangkan aturan ini berbicara tentang HUBUNGAN antar
// field — setidaknya satu di antara mereka harus ada.
func IsEmptyPatch(req model.PatchUserRequest) bool {
	return req.Username == nil && req.Email == nil && req.IsActive == nil
}
