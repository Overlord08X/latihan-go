package middleware

import (
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

// RequirePermission menolak request yang role-nya tidak memiliki
// permission tertentu. Dipasang pada route yang haknya dapat diputuskan
// TANPA melihat isi data — misalnya "boleh melihat daftar seluruh user".
//
// Untuk keputusan yang bergantung pada isi data (misalnya "boleh mengubah
// data ini karena miliknya sendiri"), pemeriksaan tidak bisa di sini:
// middleware belum tahu data siapa yang akan disentuh. Pemeriksaan
// semacam itu berada di layer service.
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			// Sampai di sini tanpa identitas berarti RequireAuth belum
			// dipasang. Tolak, jangan diloloskan.
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}
		if !perms.Can(user.Role, permission) {
			return helper.Fail(c, fiber.StatusForbidden,
				"role "+user.Role+" tidak memiliki hak "+permission)
		}
		return c.Next()
	}
}

// RequireRole memeriksa nama role secara langsung.
//
// Cara ini lebih sederhana, tetapi lebih kaku: menambah satu role baru
// berarti menyunting seluruh route yang menyebut nama role lama.
// Disediakan di sini sebagai pembanding, bukan sebagai pilihan utama.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}
		if _, granted := allowed[user.Role]; !granted {
			return helper.Fail(c, fiber.StatusForbidden,
				"role Anda tidak berhak mengakses endpoint ini")
		}
		return c.Next()
	}
}
