package helper

import (
	"github.com/gofiber/fiber/v2"

	"hunispace/app/model"
)

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func Created(c *fiber.Ctx, message string, data any) error {
	return Success(c, fiber.StatusCreated, message, data)
}
