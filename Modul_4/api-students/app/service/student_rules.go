// Package service berisi business rules murni yang tidak mengimpor Fiber sama sekali.
// Fungsi-fungsi ini menerima struct dan mengembalikan hasil — dapat diuji
// tanpa menyalakan server maupun menyentuh database.
package service

import (
	"fmt"
	"strings"

	"api-students/app/model"
)


// ValidateCreate memvalidasi request pembuatan mahasiswa baru.
// Mengembalikan Student yang siap disimpan dan map error per field.
func ValidateCreate(req model.CreateStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}
	nim := strings.TrimSpace(req.NIM)
	name := strings.TrimSpace(req.Name)

	if nim == "" {
		errs["nim"] = "nim wajib diisi"
	}
	if name == "" {
		errs["name"] = "name wajib diisi"
	}
	if req.IsActive == nil {
		errs["is_active"] = "is_active wajib diisi (true atau false)"
	}
	if err := validateGrade(req.Grade); err != nil {
		errs["grade"] = err.Error()
	}

	if len(errs) > 0 {
		return model.Student{}, errs
	}

	return model.Student{
		NIM:      nim,
		Name:     name,
		Grade:    req.Grade,
		IsActive: *req.IsActive,
	}, nil
}

// ValidateReplace memvalidasi request penggantian seluruh data mahasiswa (PUT).
// Semua field wajib hadir.
func ValidateReplace(req model.ReplaceStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.NIM == nil || strings.TrimSpace(*req.NIM) == "" {
		errs["nim"] = "nim wajib diisi"
	}
	if req.Name == nil || strings.TrimSpace(*req.Name) == "" {
		errs["name"] = "name wajib diisi"
	}
	if req.IsActive == nil {
		errs["is_active"] = "is_active wajib diisi"
	}
	if req.Grade == nil {
		errs["grade"] = "grade wajib diisi"
	} else if err := validateGrade(*req.Grade); err != nil {
		errs["grade"] = err.Error()
	}

	if len(errs) > 0 {
		return model.Student{}, errs
	}

	return model.Student{
		NIM:      strings.TrimSpace(*req.NIM),
		Name:     strings.TrimSpace(*req.Name),
		Grade:    *req.Grade,
		IsActive: *req.IsActive,
	}, nil
}

// ApplyPatch menerapkan perubahan parsial dari PatchStudentRequest ke Student yang ada.
// Hanya field yang tidak nil yang diubah.
// Mengembalikan Student baru dan map error jika ada validasi yang gagal.
func ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	errs := map[string]string{}
	result := current // salin nilai lama

	if req.NIM != nil {
		nim := strings.TrimSpace(*req.NIM)
		if nim == "" {
			errs["nim"] = "nim tidak boleh kosong"
		} else {
			result.NIM = nim
		}
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			errs["name"] = "name tidak boleh kosong"
		} else {
			result.Name = name
		}
	}
	if req.Grade != nil {
		if err := validateGrade(*req.Grade); err != nil {
			errs["grade"] = err.Error()
		} else {
			result.Grade = *req.Grade
		}
	}
	if req.IsActive != nil {
		result.IsActive = *req.IsActive
	}

	if len(errs) > 0 {
		return model.Student{}, errs
	}
	return result, nil
}

// validateGrade adalah helper internal — tidak diekspor karena hanya dipakai di service ini.
func validateGrade(grade float64) error {
	if grade < 0 || grade > 100 {
		return fmt.Errorf("grade harus antara 0 dan 100")
	}
	return nil
}
