// Package route mendaftarkan seluruh route aplikasi ke instance Fiber.
// File ini hanya berisi pemetaan URL ke method — tidak ada validasi,
// business rules, maupun query database di sini.
package route

import (
	"api-students/app/service"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterStudent mendaftarkan route untuk resource mahasiswa.
func RegisterStudent(api fiber.Router, svc *service.StudentService) {
	s := api.Group("/students", middleware.RequireJSON)
	s.Get("/", svc.List)
	s.Get("/:id", svc.Get)
	s.Post("/", svc.Create)
	s.Put("/:id", svc.Replace)
	s.Patch("/:id", svc.Patch)
	s.Delete("/:id", svc.Delete)
}
