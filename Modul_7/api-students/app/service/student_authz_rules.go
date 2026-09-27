package service

import (
	"api-students/app/model"
	"api-students/helper"
)

// CanAccessStudent memutuskan apakah seseorang boleh mengakses/mengubah data mahasiswa.
//
// Dua jalur yang diizinkan:
//  1. Kepemilikan (ownership) — ownerID sama dengan current.UserID.
//  2. Permission — role-nya memiliki permission :any yang diminta (misal: student:read:any, student:update:any).
//
// Fungsi ini adalah fungsi murni (pure function): tidak bergantung pada Fiber maupun database repository.
func CanAccessStudent(
	current model.AuthUser,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	// Jalur 1: Pemilik data selalu boleh mengakses datanya sendiri
	if current.UserID == ownerID {
		return true
	}

	// Jalur 2: Jika bukan pemilik, periksa apakah memiliki permission :any yang sesuai
	return perms.Can(current.Role, anyPermission)
}
