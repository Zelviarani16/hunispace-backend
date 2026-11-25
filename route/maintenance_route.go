package route

import (
	"github.com/gofiber/fiber/v2"

	"hunispace/app/repository"
	"hunispace/app/service"
)

func init() {
	register(func(api fiber.Router, d Deps) {
		svc := service.NewMaintenanceService(
			d.Pool,
			repository.NewLeaseRepository(),
			repository.NewMaintenanceRepository(),
		)

		tickets := api.Group("/tickets", d.Auth)
		tickets.Get("/mine", d.Tenant, svc.Mine)
		tickets.Get("", d.Owner, svc.List)
		tickets.Get("/:id", svc.Get)
		tickets.Post("", d.Tenant, svc.Create)
		tickets.Patch("/:id/status", d.Owner, svc.UpdateStatus)
	})
}
