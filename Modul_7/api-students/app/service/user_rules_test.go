package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

// TestValidateCreateUser menguji validasi deklaratif POST /users.
func TestValidateCreateUser(t *testing.T) {
	tests := []struct {
		name    string
		req     model.CreateUserRequest
		wantErr bool
		errKeys []string
	}{
		{
			name: "valid user",
			req:  model.CreateUserRequest{Username: "raihanzk", Email: "raihan@unair.ac.id", Password: "Password123!"},
		},
		{
			name:    "username terlalu pendek",
			req:     model.CreateUserRequest{Username: "ab", Email: "raihan@unair.ac.id", Password: "Password123!"},
			wantErr: true,
			errKeys: []string{"username"},
		},
		{
			name:    "email tidak valid",
			req:     model.CreateUserRequest{Username: "raihanzk", Email: "bukan-email", Password: "Password123!"},
			wantErr: true,
			errKeys: []string{"email"},
		},
		{
			name:    "password mengandung spasi (nospace)",
			req:     model.CreateUserRequest{Username: "raihanzk", Email: "raihan@unair.ac.id", Password: "pass word123"},
			wantErr: true,
			errKeys: []string{"password"},
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
						t.Errorf("diharapkan error pada field %q", key)
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

// TestValidatePatchUser menguji perilaku omitnil pada PATCH /users/:id.
func TestValidatePatchUser(t *testing.T) {
	emptyUsername := ""
	validUsername := "newuser"

	tests := []struct {
		name    string
		req     model.PatchUserRequest
		wantErr bool
		errKeys []string
	}{
		{
			name: "omitted username dan email (sah)",
			req:  model.PatchUserRequest{},
		},
		{
			name: "username valid",
			req:  model.PatchUserRequest{Username: &validUsername},
		},
		{
			name:    "username kosong DITOLAK (omitnil min=3)",
			req:     model.PatchUserRequest{Username: &emptyUsername},
			wantErr: true,
			errKeys: []string{"username"},
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
						t.Errorf("diharapkan error pada field %q", key)
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

// TestApplyPatchUser menguji ApplyPatch dan IsEmptyPatch.
func TestApplyPatchUser(t *testing.T) {
	initial := model.User{
		ID:       1,
		Username: "olduser",
		Email:    "old@unair.ac.id",
		IsActive: true,
	}

	newUsername := "updateduser"
	patched := ApplyPatch(initial, model.PatchUserRequest{Username: &newUsername})
	if patched.Username != newUsername {
		t.Errorf("expected %s, got %s", newUsername, patched.Username)
	}

	if !IsEmptyPatch(model.PatchUserRequest{}) {
		t.Error("expected IsEmptyPatch to be true for empty request")
	}
	if IsEmptyPatch(model.PatchUserRequest{Username: &newUsername}) {
		t.Error("expected IsEmptyPatch to be false")
	}
}
