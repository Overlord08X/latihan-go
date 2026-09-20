package route

import (
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterUser mendaftarkan route untuk resource users.
func RegisterUser(api fiber.Router, svc *service.UserService, perms *helper.PermissionSet) {
	users := api.Group("/users",
		middleware.RequireJSON,
		middleware.RequireAuth)

	// Hak dapat diputuskan tanpa melihat data -> middleware.
	users.Get("/",
		middleware.RequirePermission(perms, "user:list"),
		svc.List)
	users.Delete("/:id",
		middleware.RequirePermission(perms, "user:delete"),
		svc.Delete)
	users.Patch("/:id/role",
		middleware.RequirePermission(perms, "role:assign"),
		svc.AssignRole)

	// Hak bergantung pada kepemilikan data -> diperiksa di service.
	users.Get("/:id", svc.Get)
}
