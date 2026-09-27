package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

// TestValidateCreateStudent menguji validasi deklaratif input POST /students (Tugas Mandiri D.2).
func TestValidateCreateStudent(t *testing.T) {
	isActive := true
	isInactive := false

	tests := []struct {
		name    string
		req     model.CreateStudentRequest
		wantErr bool
		errKeys []string
	}{
		{
			name: "valid input",
			req:  model.CreateStudentRequest{NIM: "434241096", Name: "Raihan", Grade: 90.5, IsActive: &isActive},
		},
		{
			name:    "nim kosong",
			req:     model.CreateStudentRequest{NIM: "", Name: "Raihan", Grade: 90.0, IsActive: &isActive},
			wantErr: true,
			errKeys: []string{"nim"},
		},
		{
			name:    "nim bukan 9 digit angka (custom validation nim)",
			req:     model.CreateStudentRequest{NIM: "12345", Name: "Raihan", Grade: 90.0, IsActive: &isActive},
			wantErr: true,
			errKeys: []string{"nim"},
		},
		{
			name:    "name kosong",
			req:     model.CreateStudentRequest{NIM: "434241096", Name: "", Grade: 90.0, IsActive: &isActive},
			wantErr: true,
			errKeys: []string{"name"},
		},
		{
			name:    "is_active nil",
			req:     model.CreateStudentRequest{NIM: "434241096", Name: "Raihan", Grade: 90.0},
			wantErr: true,
			errKeys: []string{"is_active"},
		},
		{
			name:    "grade di atas 100",
			req:     model.CreateStudentRequest{NIM: "434241096", Name: "Raihan", Grade: 101.0, IsActive: &isActive},
			wantErr: true,
			errKeys: []string{"grade"},
		},
		{
			name:    "grade di bawah 0",
			req:     model.CreateStudentRequest{NIM: "434241096", Name: "Raihan", Grade: -1.0, IsActive: &isActive},
			wantErr: true,
			errKeys: []string{"grade"},
		},
		{
			name: "is_active false valid",
			req:  model.CreateStudentRequest{NIM: "434241097", Name: "Budi", Grade: 75.0, IsActive: &isInactive},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := helper.ValidateStruct(tc.req)
			if tc.wantErr {
				if len(errs) == 0 {
					t.Error("diharapkan ada error tapi tidak ada")
				}
				for _, key := range tc.errKeys {
					if _, ok := errs[key]; !ok {
						t.Errorf("diharapkan error pada field %q tapi tidak ditemukan. errors: %v", key, errs)
					}
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("tidak diharapkan ada error, tapi dapat: %v", errs)
				}
			}
		})
	}
}

// TestValidatePatchStudent menguji validasi deklaratif PATCH /students/:id dengan omitnil (Tugas Mandiri D.2).
func TestValidatePatchStudent(t *testing.T) {
	emptyName := ""
	validName := "Budi Santoso"
	invalidNIM := "abc"
	validNIM := "434241097"
	badGrade := 150.0
	goodGrade := 85.5

	tests := []struct {
		name    string
		req     model.PatchStudentRequest
		wantErr bool
		errKeys []string
	}{
		{
			name: "omitted seluruh field (sah secara tag struct)",
			req:  model.PatchStudentRequest{},
		},
		{
			name: "patch nama valid",
			req:  model.PatchStudentRequest{Name: &validName},
		},
		{
			name:    "patch nama kosong DITOLAK (omitnil tapi dikirim kosong)",
			req:     model.PatchStudentRequest{Name: &emptyName},
			wantErr: true,
			errKeys: []string{"name"},
		},
		{
			name:    "patch nim tidak valid (bukan 9 digit numerik)",
			req:     model.PatchStudentRequest{NIM: &invalidNIM},
			wantErr: true,
			errKeys: []string{"nim"},
		},
		{
			name: "patch nim valid",
			req:  model.PatchStudentRequest{NIM: &validNIM},
		},
		{
			name:    "patch grade melebihi 100",
			req:     model.PatchStudentRequest{Grade: &badGrade},
			wantErr: true,
			errKeys: []string{"grade"},
		},
		{
			name: "patch grade valid",
			req:  model.PatchStudentRequest{Grade: &goodGrade},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := helper.ValidateStruct(tc.req)
			if tc.wantErr {
				if len(errs) == 0 {
					t.Error("diharapkan ada error tapi tidak ada")
				}
				for _, key := range tc.errKeys {
					if _, ok := errs[key]; !ok {
						t.Errorf("diharapkan error pada field %q tapi tidak ditemukan. errors: %v", key, errs)
					}
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("tidak diharapkan ada error, tapi dapat: %v", errs)
				}
			}
		})
	}
}

// TestApplyPatchStudent menguji penerapan mutasi parsial pada entitas student.
func TestApplyPatchStudent(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      "434241096",
		Name:     "Raihan",
		Grade:    85.0,
		IsActive: true,
	}

	newName := "Raihan Zulfa Kamal"
	patched := ApplyPatchStudent(initial, model.PatchStudentRequest{Name: &newName})
	if patched.Name != newName {
		t.Errorf("expected name %s, got %s", newName, patched.Name)
	}
	if patched.NIM != initial.NIM {
		t.Errorf("NIM should remain %s, got %s", initial.NIM, patched.NIM)
	}

	// Uji IsEmptyPatchStudent
	if !IsEmptyPatchStudent(model.PatchStudentRequest{}) {
		t.Error("expected IsEmptyPatchStudent to be true for empty request")
	}
	if IsEmptyPatchStudent(model.PatchStudentRequest{Name: &newName}) {
		t.Error("expected IsEmptyPatchStudent to be false when name is provided")
	}
}
