package route

import (
	"api-students/app/service"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterReview(api fiber.Router, svc *service.ReviewService) {
	// --- ulasan kelas bersarang di /courses ---
	courses := api.Group("/courses", middleware.RequireJSON, middleware.RequireAuth)
	courses.Get("/:id/reviews", svc.ListByCourse)
	courses.Get("/:id/reviews/summary", svc.Summary)
	courses.Post("/:id/reviews", svc.Create)

	// --- endpoint ulasan mandiri di /reviews ---
	reviews := api.Group("/reviews", middleware.RequireJSON, middleware.RequireAuth)
	reviews.Get("/:id", svc.Get)
	reviews.Patch("/:id", svc.Update)
	reviews.Delete("/:id", svc.Delete)
	reviews.Post("/:id/replies", svc.Reply)
	reviews.Post("/:id/helpful", svc.MarkHelpful)
}
