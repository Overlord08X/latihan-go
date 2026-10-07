package route

import (
	"siakad-mini/app/model"
	"siakad-mini/app/service"
	"siakad-mini/helper"
	"siakad-mini/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterCourse(r fiber.Router, courseService *service.CourseService, jwtSecret string) {
	courses := r.Group("/courses", middleware.RequireAuth(jwtSecret))

	// Endpoint 8: GET /api/v1/courses (Semua role terotentikasi)
	courses.Get("/", func(c *fiber.Ctx) error {
		var f model.CourseFilterQuery
		if err := c.QueryParser(&f); err != nil {
			return helper.UnprocessableEntity("Parameter query tidak valid", nil)
		}

		list, err := courseService.List(c.Context(), f)
		if err != nil {
			return err
		}

		return helper.Success(c, fiber.StatusOK, "Daftar mata kuliah berhasil diambil", list)
	})
}
