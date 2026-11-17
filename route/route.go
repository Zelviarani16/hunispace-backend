package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"hunispace/app/repository"
	"hunispace/helper"
)

type Deps struct {
	Pool *pgxpool.Pool
	Tx   *repository.TxManager

	Auth   func(*fiber.Ctx) error
	Owner  func(*fiber.Ctx) error
	Tenant func(*fiber.Ctx) error

	JWT *helper.JWTManager
}

type registrar func(api fiber.Router, d Deps)

var registrars []registrar

func register(r registrar) {
	registrars = append(registrars, r)
}

func Setup(app *fiber.App, d Deps) {
	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := d.Pool.Ping(ctx); err != nil {
			return helper.NewError(fiber.StatusServiceUnavailable,
				helper.CodeServiceUnavailable, "database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "HuniSpace berjalan", fiber.Map{"database": "ok"})
	})

	api := app.Group("/api/v1")
	for _, r := range registrars {
		r(api, d)
	}
}
