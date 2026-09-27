package route

import (
	"api-students/app/service"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

// RegisterNilai mendaftarkan route untuk resource nilai.
// Seluruh endpoint nilai dilindungi RequireAuth.
func RegisterNilai(api fiber.Router, svc *service.NilaiService) {
	// Group untuk pembuatan nilai — dilindungi RequireAuth
	n := api.Group("/nilais", middleware.RequireAuth, middleware.RequireJSON)
	n.Post("/", svc.Create)

	// Endpoint untuk melihat nilai mahasiswa berdasarkan NIM — dilindungi RequireAuth
	api.Get("/students/:nim/nilais", middleware.RequireAuth, svc.GetByStudentNIM)
}
