package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"hunispace/app/model"
	"hunispace/app/repository"
	"hunispace/helper"
)

type MaintenanceService struct {
	pool    *pgxpool.Pool
	leases  repository.LeaseRepository
	tickets repository.MaintenanceRepository
}

func NewMaintenanceService(
	pool *pgxpool.Pool,
	leases repository.LeaseRepository,
	tickets repository.MaintenanceRepository,
) *MaintenanceService {
	return &MaintenanceService{pool: pool, leases: leases, tickets: tickets}
}

func (s *MaintenanceService) Create(c *fiber.Ctx) error {
	var req model.TicketCreateRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	if fields := validateTicketCreate(&req); len(fields) > 0 {
		return helper.Validation(fields)
	}

	ctx := c.UserContext()
	auth := helper.CurrentUser(c)

	lease, err := s.leases.FindActiveByTenant(ctx, s.pool, auth.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Conflict("NO_ACTIVE_LEASE", "anda tidak memiliki kontrak sewa aktif")
		}
		return helper.Internal(err)
	}

	ticket, err := s.tickets.Create(ctx, s.pool, model.Ticket{
		TenantID:    auth.UserID,
		RoomID:      lease.RoomID,
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Priority:    req.Priority,
	})
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Created(c, "keluhan berhasil dicatat", ticket)
}

func (s *MaintenanceService) Mine(c *fiber.Ctx) error {
	auth := helper.CurrentUser(c)
	tickets, err := s.tickets.FindByTenant(c.UserContext(), s.pool, auth.UserID)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "keluhan saya berhasil diambil", tickets)
}

func (s *MaintenanceService) List(c *fiber.Ctx) error {
	status := c.Query("status")
	if status != "" && !validTicketStatus(status) {
		return helper.BadRequest("status harus open, in_progress, atau resolved")
	}
	tickets, err := s.tickets.FindAll(c.UserContext(), s.pool, status)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar keluhan berhasil diambil", tickets)
}

func (s *MaintenanceService) UpdateStatus(c *fiber.Ctx) error {
	id, err := helper.ParseID(c, "id")
	if err != nil {
		return err
	}
	var req model.TicketStatusRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	if fields := validateTicketStatus(req); len(fields) > 0 {
		return helper.Validation(fields)
	}

	ctx := c.UserContext()
	ticket, err := s.tickets.UpdateStatus(ctx, s.pool, id, req.Status, req.Version)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return helper.Internal(err)
		}
		if _, findErr := s.tickets.FindByID(ctx, s.pool, id); findErr != nil {
			return lookupError(findErr, "tiket")
		}
		return helper.Conflict("VERSION_CONFLICT",
			"tiket sudah diubah pihak lain; muat ulang tiket lalu coba lagi")
	}
	return helper.Success(c, fiber.StatusOK, "status tiket berhasil diperbarui", ticket)
}
