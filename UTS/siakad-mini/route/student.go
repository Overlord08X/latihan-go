package route

import (
	"strconv"

	"siakad-mini/app/model"
	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterStudent(r fiber.Router, studentService *service.StudentService, jwtSecret string) {
	students := r.Group("/students", middleware.RequireAuth(jwtSecret))

	// Endpoint 3: GET /api/v1/students (Khusus Admin)
	students.Get("/", middleware.RequireRole("admin"), func(c *fiber.Ctx) error {
		var f model.StudentFilterQuery
		if err := c.QueryParser(&f); err != nil {
			return helper.UnprocessableEntity("Parameter query tidak valid", nil)
		}

		list, meta, err := studentService.List(c.Context(), f)
		if err != nil {
			return err
		}

		return helper.SuccessWithMeta(c, fiber.StatusOK, "Data mahasiswa berhasil diambil", list, meta)
	})

	// Endpoint 4: POST /api/v1/students (Khusus Admin)
	students.Post("/", middleware.RequireRole("admin"), func(c *fiber.Ctx) error {
		var req model.CreateStudentRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.UnprocessableEntity("Format JSON tidak valid", map[string][]string{
				"body": {"Body request harus berformat JSON yang valid"},
			})
		}

		created, err := studentService.Create(c.Context(), req)
		if err != nil {
			return err
		}

		return helper.Success(c, fiber.StatusCreated, "Mahasiswa dan akun berhasil dibuat", created)
	})

	// Endpoint 5: GET /api/v1/students/:id (Admin, Mahasiswa data sendiri)
	students.Get("/:id", func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}

		role := c.Locals("role").(string)
		studentID := c.Locals("studentID").(int64)

		detail, err := studentService.Get(c.Context(), role, studentID, id)
		if err != nil {
			return err
		}

		return helper.Success(c, fiber.StatusOK, "Detail mahasiswa berhasil diambil", detail)
	})

	// Endpoint 6: PUT /api/v1/students/:id (Khusus Admin)
	students.Put("/:id", middleware.RequireRole("admin"), func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}

		var req model.UpdateStudentRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.UnprocessableEntity("Format JSON tidak valid", map[string][]string{
				"body": {"Body request harus berformat JSON yang valid"},
			})
		}

		updated, err := studentService.Update(c.Context(), id, req)
		if err != nil {
			return err
		}

		return helper.Success(c, fiber.StatusOK, "Data mahasiswa berhasil diperbarui", updated)
	})

	// Endpoint 7: DELETE /api/v1/students/:id (Khusus Admin)
	students.Delete("/:id", middleware.RequireRole("admin"), func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return helper.NotFound("Mahasiswa tidak ditemukan")
		}

		if err := studentService.Delete(c.Context(), id); err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusNoContent)
	})
}
