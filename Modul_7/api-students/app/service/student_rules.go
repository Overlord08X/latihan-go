package service

import (
	"strings"

	"api-students/app/model"
)

// ApplyPatchStudent menerapkan perubahan parsial dari PatchStudentRequest ke Student yang ada (Tugas Mandiri D.2).
func ApplyPatchStudent(current model.Student, req model.PatchStudentRequest) model.Student {
	if req.NIM != nil {
		current.NIM = strings.TrimSpace(*req.NIM)
	}
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current
}

// IsEmptyPatchStudent memeriksa body PATCH students yang tidak berisi field apa pun.
func IsEmptyPatchStudent(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}
