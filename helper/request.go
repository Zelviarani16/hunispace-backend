package helper

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"hunispace/app/model"
)

const LocalsAuthUser = "auth_user"

func CurrentUser(c *fiber.Ctx) model.AuthUser {
	user, _ := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user
}

func ParseID(c *fiber.Ctx, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, BadRequest("parameter " + name + " harus berupa angka positif")
	}
	return id, nil
}

func ParseBody(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return BadRequest("body request bukan JSON yang valid")
	}
	return nil
}
