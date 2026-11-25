package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"hunispace/config"
)

func SetupSwagger(app *fiber.App) {
	if config.GetEnv("SWAGGER_ENABLED", "false") != "true" {
		return
	}
	app.Get("/swagger/*", swagger.HandlerDefault)
}
