package route

import (
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterSupervisor(api fiber.Router, svc *service.SupervisorService, perms *helper.PermissionSet) {
	// --- dosen pembimbing dan bimbingan tugas akhir ---
	supervisors := api.Group("/supervisors",
		middleware.RequireJSON,
		middleware.RequireAuth)
	supervisors.Get("/", svc.List)
	supervisors.Get("/:id", svc.Get)
	supervisors.Post("/:id/assign", middleware.RequirePermission(perms, "supervision:manage"), svc.Assign)
	supervisors.Patch("/:id", middleware.RequirePermission(perms, "supervision:manage"), svc.UpdateQuota)

	supervisions := api.Group("/supervisions",
		middleware.RequireJSON,
		middleware.RequireAuth)
	supervisions.Post("/switch", middleware.RequirePermission(perms, "supervision:manage"), svc.Switch)
}
