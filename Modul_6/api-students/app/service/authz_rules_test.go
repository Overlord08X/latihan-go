package service

import (
	"testing"

	"api-students/app/model"
	"api-students/helper"
)

func createTestPermissionSet() *helper.PermissionSet {
	raw := map[string][]string{
		"admin": {
			"user:list", "user:read:any", "user:update:any", "user:delete", "role:assign",
			"student:list", "student:read:any", "student:create", "student:update:any", "student:delete",
		},
		"staff": {
			"user:list", "user:read:any",
			"student:list", "student:read:any", "student:create",
		},
		"user": {},
	}
	return helper.NewPermissionSet(raw)
}

func TestCanAccessUser(t *testing.T) {
	perms := createTestPermissionSet()

	admin := model.AuthUser{UserID: 1, Username: "admin", Role: "admin"}
	staff := model.AuthUser{UserID: 2, Username: "staff", Role: "staff"}
	user := model.AuthUser{UserID: 3, Username: "user", Role: "user"}

	tests := []struct {
		name          string
		current       model.AuthUser
		targetID      int
		anyPermission string
		want          bool
	}{
		{"User akses diri sendiri", user, 3, "user:read:any", true},
		{"User akses orang lain", user, 2, "user:read:any", false},
		{"Staff akses diri sendiri", staff, 2, "user:read:any", true},
		{"Staff akses orang lain dengan user:read:any", staff, 3, "user:read:any", true},
		{"Staff update orang lain tanpa user:update:any", staff, 3, "user:update:any", false},
		{"Admin update orang lain dengan user:update:any", admin, 3, "user:update:any", true},
		{"Role tidak dikenal", model.AuthUser{UserID: 99, Role: "guest"}, 99, "user:read:any", true},
		{"Role tidak dikenal akses orang lain", model.AuthUser{UserID: 99, Role: "guest"}, 1, "user:read:any", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanAccessUser(tt.current, tt.targetID, perms, tt.anyPermission)
			if got != tt.want {
				t.Errorf("CanAccessUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAssignRole(t *testing.T) {
	perms := createTestPermissionSet()
	admin := model.AuthUser{UserID: 1, Username: "admin", Role: "admin"}

	t.Run("Valid assign role to other user", func(t *testing.T) {
		errs := ValidateAssignRole(admin, 2, model.AssignRoleRequest{Role: "staff"}, perms)
		if len(errs) != 0 {
			t.Errorf("expected no errors, got %v", errs)
		}
	})

	t.Run("Error assign role to self", func(t *testing.T) {
		errs := ValidateAssignRole(admin, 1, model.AssignRoleRequest{Role: "user"}, perms)
		if errs["role"] != "tidak boleh mengubah role diri sendiri" {
			t.Errorf("expected self role change error, got %v", errs)
		}
	})

	t.Run("Error unknown role", func(t *testing.T) {
		errs := ValidateAssignRole(admin, 2, model.AssignRoleRequest{Role: "superadmin"}, perms)
		if errs["role"] == "" {
			t.Errorf("expected unknown role error, got %v", errs)
		}
	})

	t.Run("Error empty role", func(t *testing.T) {
		errs := ValidateAssignRole(admin, 2, model.AssignRoleRequest{Role: ""}, perms)
		if errs["role"] != "wajib diisi" {
			t.Errorf("expected required role error, got %v", errs)
		}
	})
}

func TestCanAccessStudent(t *testing.T) {
	perms := createTestPermissionSet()

	admin := model.AuthUser{UserID: 1, Username: "admin", Role: "admin"}
	staff := model.AuthUser{UserID: 2, Username: "staff", Role: "staff"}
	userOwner := model.AuthUser{UserID: 3, Username: "sari", Role: "user"}
	userOther := model.AuthUser{UserID: 4, Username: "budi", Role: "user"}

	tests := []struct {
		name          string
		current       model.AuthUser
		ownerID       int
		anyPermission string
		want          bool
	}{
		{"User pemilik boleh read datanya sendiri", userOwner, 3, "student:read:any", true},
		{"User pemilik boleh update datanya sendiri", userOwner, 3, "student:update:any", true},
		{"User non-pemilik ditolak read data orang lain", userOther, 3, "student:read:any", false},
		{"User non-pemilik ditolak update data orang lain", userOther, 3, "student:update:any", false},
		{"Staff boleh read data orang lain dengan student:read:any", staff, 3, "student:read:any", true},
		{"Staff ditolak update data orang lain tanpa student:update:any", staff, 3, "student:update:any", false},
		{"Admin boleh read data siapapun dengan student:read:any", admin, 3, "student:read:any", true},
		{"Admin boleh update data siapapun dengan student:update:any", admin, 3, "student:update:any", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanAccessStudent(tt.current, tt.ownerID, perms, tt.anyPermission)
			if got != tt.want {
				t.Errorf("CanAccessStudent() = %v, want %v", got, tt.want)
			}
		})
	}
}
