package route

import (
	"github.com/gofiber/fiber/v2"

	"hunispace/app/repository"
	"hunispace/app/service"
)

func init() {
	register(func(api fiber.Router, d Deps) {
		svc := service.NewUserService(d.Pool, repository.NewUserRepository(), d.JWT)

		auth := api.Group("/auth")
		auth.Post("/register", svc.Register)
		auth.Post("/login", svc.Login)
		auth.Get("/me", d.Auth, svc.Me)
	})
}
