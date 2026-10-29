package route

import (
	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

func RegisterNotification(api fiber.Router, svc *service.NotificationService, perms *helper.PermissionSet) {
	// --- notifikasi (diisi oleh worker, dibaca lewat API) ---
	notifications := api.Group("/notifications",
		middleware.RequireJSON,
		middleware.RequireAuth)
	notifications.Get("/mine", svc.Mine)

	admin := api.Group("/admin",
		middleware.RequireJSON,
		middleware.RequireAuth)
	admin.Get("/queue", middleware.RequirePermission(perms, "queue:monitor"),
		svc.Monitor)
}
