package service

import (
	"testing"

	"api-students/app/model"
)

func TestValidateCreateNilai(t *testing.T) {
	idValid := 1
	idInvalid := 0
	nilaiValid := 90.5
	nilaiInvalidMax := 105.0
	nilaiInvalidMin := -5.0

	tests := []struct {
		name    string
		req     *model.CreateNilaiRequest
		wantErr bool
		errKey  string
	}{
		{
			name: "valid input",
			req: &model.CreateNilaiRequest{
				IDStudent:  &idValid,
				NamaMatkul: "Matematika",
				Nilai:      &nilaiValid,
			},
			wantErr: false,
		},
		{
			name: "id_student kosong",
			req: &model.CreateNilaiRequest{
				IDStudent:  nil,
				NamaMatkul: "Matematika",
				Nilai:      &nilaiValid,
			},
			wantErr: true,
			errKey:  "id_student",
		},
		{
			name: "id_student nol atau negatif",
			req: &model.CreateNilaiRequest{
				IDStudent:  &idInvalid,
				NamaMatkul: "Matematika",
				Nilai:      &nilaiValid,
			},
			wantErr: true,
			errKey:  "id_student",
		},
		{
			name: "nama_matkul kosong",
			req: &model.CreateNilaiRequest{
				IDStudent:  &idValid,
				NamaMatkul: "   ",
				Nilai:      &nilaiValid,
			},
			wantErr: true,
			errKey:  "nama_matkul",
		},
		{
			name: "nilai kosong",
			req: &model.CreateNilaiRequest{
				IDStudent:  &idValid,
				NamaMatkul: "Matematika",
				Nilai:      nil,
			},
			wantErr: true,
			errKey:  "nilai",
		},
		{
			name: "nilai lebih dari 100",
			req: &model.CreateNilaiRequest{
				IDStudent:  &idValid,
				NamaMatkul: "Matematika",
				Nilai:      &nilaiInvalidMax,
			},
			wantErr: true,
			errKey:  "nilai",
		},
		{
			name: "nilai kurang dari 0",
			req: &model.CreateNilaiRequest{
				IDStudent:  &idValid,
				NamaMatkul: "Matematika",
				Nilai:      &nilaiInvalidMin,
			},
			wantErr: true,
			errKey:  "nilai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := ValidateCreateNilai(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateNilai() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errKey != "" {
				if _, ok := errs[tt.errKey]; !ok {
					t.Errorf("ValidateCreateNilai() errs = %v, harus mengandung key %v", errs, tt.errKey)
				}
			}
		})
	}
}
