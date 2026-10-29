package route

import (
	"api-students/app/service"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterPortfolio(api fiber.Router, svc *service.PortfolioService) {
	// --- portofolio mahasiswa (Tugas Mandiri C) ---
	students := api.Group("/students", middleware.RequireJSON, middleware.RequireAuth)
	students.Get("/:id/portfolio", svc.Get)
	students.Put("/:id/portfolio", svc.Upsert)
	students.Post("/:id/portfolio/projects", svc.AddProject)
	students.Delete("/:id/portfolio/projects/:projectId", svc.DeleteProject)

	portfolios := api.Group("/portfolios", middleware.RequireJSON, middleware.RequireAuth)
	portfolios.Get("/stats/technologies", svc.TechnologyStats)
}
