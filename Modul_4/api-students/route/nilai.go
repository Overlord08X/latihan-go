package route

import (
	"api-students/app/service"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterNilai mendaftarkan route untuk resource nilai.
func RegisterNilai(api fiber.Router, svc *service.NilaiService) {
	// Group untuk pembuatan nilai
	n := api.Group("/nilais", middleware.RequireJSON)
	n.Post("/", svc.Create)

	// Endpoint untuk melihat nilai mahasiswa berdasarkan NIM
	api.Get("/students/:nim/nilais", svc.GetByStudentNIM)
}
