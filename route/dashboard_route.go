package route

import (
	"github.com/gofiber/fiber/v2"

	"hunispace/app/repository"
	"hunispace/app/service"
)

func init() {
	register(func(api fiber.Router, d Deps) {
		svc := service.NewDashboardService(d.Pool, repository.NewDashboardRepository())
		api.Get("/dashboard/summary", d.Auth, d.Owner, svc.Summary)
	})
}
