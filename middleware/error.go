package middleware

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

	"hunispace/app/model"
	"hunispace/helper"
)

func ErrorResponse() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		var appErr *helper.AppError
		if errors.As(err, &appErr) {
			if appErr.Status >= fiber.StatusInternalServerError {
				log.Printf("error server: %v", err)
			}
			return c.Status(appErr.Status).JSON(model.WebResponse{
				Message: appErr.Message,
				Error: &model.ErrorBody{
					Code: appErr.Code, Message: appErr.Message, Fields: appErr.Fields,
				},
			})
		}

		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			code := helper.CodeBadRequest
			if fiberErr.Code == fiber.StatusNotFound {
				code = helper.CodeNotFound
			}
			return c.Status(fiberErr.Code).JSON(model.WebResponse{
				Message: fiberErr.Message,
				Error:   &model.ErrorBody{Code: code, Message: fiberErr.Message},
			})
		}

		log.Printf("error tidak terduga: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(model.WebResponse{
			Message: "terjadi kesalahan pada server",
			Error: &model.ErrorBody{
				Code: helper.CodeInternal, Message: "terjadi kesalahan pada server",
			},
		})
	}
}
