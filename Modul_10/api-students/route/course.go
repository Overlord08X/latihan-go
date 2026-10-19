package route

import (
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterCourse(api fiber.Router, svc *service.CourseService, perms *helper.PermissionSet) {
	// --- kelas dan KRS ---
	courses := api.Group("/courses",
		middleware.RequireJSON,
		middleware.RequireAuth)

	// "/mine" didaftarkan SEBELUM "/:id". Fiber mencocokkan route sesuai
	// urutan pendaftaran; bila dibalik, "mine" akan dianggap sebagai :id.
	courses.Get("/", svc.List)
	courses.Get("/mine", svc.Mine)
	courses.Get("/:id", svc.Get)
	courses.Post("/:id/enroll", svc.Enroll)
	courses.Delete("/:id/enroll", svc.Drop)
	courses.Patch("/:id", middleware.RequirePermission(perms, "course:manage"), svc.UpdateQuota)

	enrollments := api.Group("/enrollments",
		middleware.RequireJSON,
		middleware.RequireAuth)
	enrollments.Post("/swap", svc.Swap)
}
