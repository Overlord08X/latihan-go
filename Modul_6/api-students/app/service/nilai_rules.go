package service

import (
	"errors"
	"strings"

	"api-students/app/model"
)

// ValidateCreateNilai memvalidasi input pembuatan Nilai tanpa bergantung pada HTTP context.
// Mengembalikan error dengan daftar pesan (map[string]string) jika ada field yang salah.
func ValidateCreateNilai(req *model.CreateNilaiRequest) (map[string]string, error) {
	errs := make(map[string]string)

	if req.IDStudent == nil || *req.IDStudent <= 0 {
		errs["id_student"] = "id_student wajib diisi dan harus berupa angka positif"
	}

	nama := strings.TrimSpace(req.NamaMatkul)
	if nama == "" {
		errs["nama_matkul"] = "nama_matkul wajib diisi"
	}

	if req.Nilai == nil {
		errs["nilai"] = "nilai wajib diisi"
	} else if *req.Nilai < 0 || *req.Nilai > 100 {
		errs["nilai"] = "nilai harus berada di antara 0.00 hingga 100.00"
	}

	if len(errs) > 0 {
		return errs, errors.New("validasi gagal")
	}

	return nil, nil
}
