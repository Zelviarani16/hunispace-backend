package middleware

import (
	"errors"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"

	"hunispace/helper"
)

func RequireAuth(jwtManager *helper.JWTManager) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			return helper.Unauthorized("header Authorization tidak ada atau salah bentuk")
		}

		authUser, err := jwtManager.Parse(token)
		if err != nil {
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Unauthorized("access token kedaluwarsa")
			}
			return helper.Unauthorized("access token tidak valid")
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

func RequireRole(roles ...string) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user := helper.CurrentUser(c)
		if !slices.Contains(roles, user.Role) {
			return helper.Forbidden("anda tidak berhak mengakses endpoint ini")
		}
		return c.Next()
	}
}

func bearerToken(c *fiber.Ctx) (string, error) {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", errors.New("header kosong")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("format bukan Bearer")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token kosong")
	}
	return token, nil
}
