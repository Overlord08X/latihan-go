package service

import (
	"testing"

	"api-students/app/model"
)

// Perhatikan: pengujian ini tidak menyalakan server, tidak menyentuh
// database, dan tidak membuat fiber.Ctx.

// ---------------------------------------------------------------------------
// TestCountTotalPages — helper function (dipindah ke CountTotalPages di helper)
// ---------------------------------------------------------------------------

// TestValidateCreate menguji validasi input POST /students.
func TestValidateCreate(t *testing.T) {
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
			name:    "semua field kosong",
			req:     model.CreateStudentRequest{},
			wantErr: true,
			errKeys: []string{"nim", "name", "is_active"},
		},
		{
			name: "is_active false valid",
			req:  model.CreateStudentRequest{NIM: "434241097", Name: "Budi", Grade: 75.0, IsActive: &isInactive},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			student, errs := ValidateCreate(tc.req)
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
				if student.NIM == "" {
					t.Error("student.NIM tidak boleh kosong setelah validasi sukses")
				}
			}
		})
	}
}

// TestValidateReplace menguji validasi input PUT /students/:id.
func TestValidateReplace(t *testing.T) {
	nim := "434241096"
	name := "Raihan"
	grade := 88.5
	isActive := true
	badGrade := 150.0

	tests := []struct {
		name    string
		req     model.ReplaceStudentRequest
		wantErr bool
		errKeys []string
	}{
		{
			name: "valid input",
			req:  model.ReplaceStudentRequest{NIM: &nim, Name: &name, Grade: &grade, IsActive: &isActive},
		},
		{
			name:    "nim nil",
			req:     model.ReplaceStudentRequest{Name: &name, Grade: &grade, IsActive: &isActive},
			wantErr: true,
			errKeys: []string{"nim"},
		},
		{
			name:    "grade nil",
			req:     model.ReplaceStudentRequest{NIM: &nim, Name: &name, IsActive: &isActive},
			wantErr: true,
			errKeys: []string{"grade"},
		},
		{
			name:    "grade tidak valid",
			req:     model.ReplaceStudentRequest{NIM: &nim, Name: &name, Grade: &badGrade, IsActive: &isActive},
			wantErr: true,
			errKeys: []string{"grade"},
		},
		{
			name:    "semua nil",
			req:     model.ReplaceStudentRequest{},
			wantErr: true,
			errKeys: []string{"nim", "name", "grade", "is_active"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			student, errs := ValidateReplace(tc.req)
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
				if student.Name == "" {
					t.Error("student.Name tidak boleh kosong setelah validasi sukses")
				}
			}
		})
	}
}

// TestApplyPatch menguji penerapan perubahan parsial PATCH /students/:id.
func TestApplyPatch(t *testing.T) {
	initial := model.Student{
		ID:       1,
		NIM:      "434241096",
		Name:     "Raihan Zaky",
		Grade:    90.5,
		IsActive: true,
	}

	isInactive := false
	newGrade := 95.0
	newName := "Raihan Zaky Updated"
	emptyName := ""
	badGrade := -5.0

	tests := []struct {
		name        string
		req         model.PatchStudentRequest
		wantErr     bool
		checkResult func(t *testing.T, s model.Student)
	}{
		{
			name: "patch is_active saja",
			req:  model.PatchStudentRequest{IsActive: &isInactive},
			checkResult: func(t *testing.T, s model.Student) {
				if s.IsActive {
					t.Error("is_active seharusnya berubah menjadi false")
				}
				if s.NIM != initial.NIM {
					t.Error("nim seharusnya tidak berubah")
				}
				if s.Name != initial.Name {
					t.Error("name seharusnya tidak berubah")
				}
			},
		},
		{
			name: "patch grade saja",
			req:  model.PatchStudentRequest{Grade: &newGrade},
			checkResult: func(t *testing.T, s model.Student) {
				if s.Grade != 95.0 {
					t.Errorf("grade seharusnya 95.0, dapat %v", s.Grade)
				}
				if s.Name != initial.Name {
					t.Error("name seharusnya tidak berubah")
				}
			},
		},
		{
			name: "patch name saja",
			req:  model.PatchStudentRequest{Name: &newName},
			checkResult: func(t *testing.T, s model.Student) {
				if s.Name != "Raihan Zaky Updated" {
					t.Errorf("name seharusnya berubah, dapat %q", s.Name)
				}
			},
		},
		{
			name:    "name kosong → error",
			req:     model.PatchStudentRequest{Name: &emptyName},
			wantErr: true,
		},
		{
			name:    "grade tidak valid → error",
			req:     model.PatchStudentRequest{Grade: &badGrade},
			wantErr: true,
		},
		{
			name: "tidak ada field → tidak ada perubahan, tidak ada error",
			req:  model.PatchStudentRequest{},
			checkResult: func(t *testing.T, s model.Student) {
				if s.NIM != initial.NIM || s.Name != initial.Name || s.Grade != initial.Grade {
					t.Error("tidak ada field yang berubah")
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, errs := ApplyPatch(initial, tc.req)
			if tc.wantErr {
				if len(errs) == 0 {
					t.Error("diharapkan ada error tapi tidak ada")
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("tidak diharapkan ada error, tapi dapat: %v", errs)
				}
				if tc.checkResult != nil {
					tc.checkResult(t, result)
				}
			}
		})
	}
}
