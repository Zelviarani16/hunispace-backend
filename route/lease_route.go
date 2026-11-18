package route

import (
	"github.com/gofiber/fiber/v2"

	"hunispace/app/repository"
	"hunispace/app/service"
)

func init() {
	register(func(api fiber.Router, d Deps) {
		svc := service.NewLeaseService(
			d.Pool, d.Tx,
			repository.NewUserRepository(),
			repository.NewRoomRepository(),
			repository.NewLeaseRepository(),
		)

		leases := api.Group("/leases", d.Auth)
		leases.Get("/mine", d.Tenant, svc.Mine)
		leases.Get("", d.Owner, svc.List)
		leases.Post("", d.Owner, svc.Create)
		leases.Post("/:id/end", d.Owner, svc.End)
	})
}
