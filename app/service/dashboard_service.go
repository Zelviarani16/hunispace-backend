package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"hunispace/app/repository"
	"hunispace/helper"
)

type DashboardService struct {
	pool      *pgxpool.Pool
	dashboard repository.DashboardRepository
}

func NewDashboardService(pool *pgxpool.Pool, dashboard repository.DashboardRepository) *DashboardService {
	return &DashboardService{pool: pool, dashboard: dashboard}
}

func (s *DashboardService) Summary(c *fiber.Ctx) error {
	summary, err := s.dashboard.Summary(c.UserContext(), s.pool)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "ringkasan berhasil diambil", summary)
}
