// Package route mendaftarkan seluruh route aplikasi ke instance Fiber.
package route

import (
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterStudent mendaftarkan route untuk resource mahasiswa.
func RegisterStudent(api fiber.Router, svc *service.StudentService, perms *helper.PermissionSet) {
	s := api.Group("/students", middleware.RequireAuth, middleware.RequireJSON)

	// Hak dapat diputuskan tanpa melihat data -> middleware
	s.Get("/", middleware.RequirePermission(perms, "student:list"), svc.List)
	s.Post("/", middleware.RequirePermission(perms, "student:create"), svc.Create)
	s.Delete("/:id", middleware.RequirePermission(perms, "student:delete"), svc.Delete)

	// Hak bergantung pada kepemilikan data -> diperiksa di service
	s.Get("/:id", svc.Get)
	s.Put("/:id", svc.Replace)
	s.Patch("/:id", svc.Patch)
}
