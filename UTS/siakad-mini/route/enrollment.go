package route

import (
	"strconv"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterEnrollment(r fiber.Router, enrollmentService *service.EnrollmentService, jwtSecret string) {
	enrollments := r.Group("/enrollments", middleware.RequireAuth(jwtSecret), middleware.RequireRole("mahasiswa"))

	// Endpoint 9: POST /api/v1/enrollments (Khusus Mahasiswa)
	enrollments.Post("/", func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(int64)

		var req model.CreateEnrollmentRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.UnprocessableEntity("Format JSON tidak valid", map[string][]string{
				"body": {"Body request harus berformat JSON yang valid"},
			})
		}

		created, err := enrollmentService.Create(c.Context(), userID, req)
		if err != nil {
			return err
		}

		return helper.Success(c, fiber.StatusCreated, "Mata kuliah berhasil ditambahkan ke KRS", created)
	})

	// Endpoint 10: DELETE /api/v1/enrollments/:id (Khusus Mahasiswa milik sendiri)
	enrollments.Delete("/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return helper.NotFound("Data KRS tidak ditemukan")
		}

		userID := c.Locals("userID").(int64)

		if err := enrollmentService.Delete(c.Context(), userID, id); err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}
