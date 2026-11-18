package route

import (
	"github.com/gofiber/fiber/v2"

	"hunispace/app/repository"
	"hunispace/app/service"
)

func init() {
	register(func(api fiber.Router, d Deps) {
		svc := service.NewRoomService(d.Pool, repository.NewRoomRepository())

		rooms := api.Group("/rooms", d.Auth)
		rooms.Get("", svc.List)
		rooms.Get("/:id", svc.Get)
		rooms.Post("", d.Owner, svc.Create)
		rooms.Patch("/:id", d.Owner, svc.Update)
		rooms.Delete("/:id", d.Owner, svc.Delete)
	})
}
