package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"hunispace/app/model"
	"hunispace/app/repository"
	"hunispace/helper"
)

type LeaseService struct {
	pool   *pgxpool.Pool
	tx     *repository.TxManager
	users  repository.UserRepository
	rooms  repository.RoomRepository
	leases repository.LeaseRepository
}

func NewLeaseService(
	pool *pgxpool.Pool,
	tx *repository.TxManager,
	users repository.UserRepository,
	rooms repository.RoomRepository,
	leases repository.LeaseRepository,
) *LeaseService {
	return &LeaseService{pool: pool, tx: tx, users: users, rooms: rooms, leases: leases}
}

func (s *LeaseService) Create(c *fiber.Ctx) error {
	var req model.LeaseCreateRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	start, end, fields := ValidateLeaseRequest(req)
	if len(fields) > 0 {
		return helper.Validation(fields)
	}

	ctx := c.UserContext()

	tenant, err := s.users.FindByID(ctx, s.pool, req.TenantID)
	if err != nil {
		return lookupError(err, "penghuni")
	}
	if tenant.Role != model.RoleTenant {
		return helper.BadRequest("kontrak hanya dapat dibuat untuk akun berperan tenant")
	}

	var lease model.Lease
	err = s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		room, err := s.rooms.FindByIDForUpdate(ctx, tx, req.RoomID)
		if err != nil {
			return lookupError(err, "kamar")
		}
		if room.Status != model.RoomAvailable {
			return helper.Conflict("ROOM_NOT_AVAILABLE", "kamar tidak tersedia (status: "+room.Status+")")
		}

		lease, err = s.leases.Create(ctx, tx, model.Lease{
			TenantID: req.TenantID, RoomID: req.RoomID, StartDate: start, EndDate: end,
		})
		if err != nil {
			if errors.Is(err, repository.ErrDuplicate) {
				return helper.Conflict("ROOM_NOT_AVAILABLE", "kamar sudah memiliki kontrak aktif")
			}
			return helper.Internal(err)
		}

		if err := s.rooms.UpdateStatus(ctx, tx, req.RoomID, model.RoomOccupied); err != nil {
			return helper.Internal(err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return helper.Created(c, "kontrak sewa berhasil dibuat", lease)
}

func (s *LeaseService) End(c *fiber.Ctx) error {
	id, err := helper.ParseID(c, "id")
	if err != nil {
		return err
	}
	ctx := c.UserContext()

	var lease model.Lease
	err = s.tx.WithTx(ctx, func(tx repository.DBTX) error {
		current, err := s.leases.FindByIDForUpdate(ctx, tx, id)
		if err != nil {
			return lookupError(err, "kontrak")
		}
		if current.Status != model.LeaseActive {
			return helper.Conflict("LEASE_ALREADY_ENDED", "kontrak sudah berakhir")
		}

		if err := s.leases.SetStatus(ctx, tx, id, model.LeaseEnded); err != nil {
			return helper.Internal(err)
		}
		if err := s.rooms.UpdateStatus(ctx, tx, current.RoomID, model.RoomAvailable); err != nil {
			return helper.Internal(err)
		}

		current.Status = model.LeaseEnded
		lease = current
		return nil
	})
	if err != nil {
		return err
	}
	return helper.Success(c, fiber.StatusOK, "kontrak sewa berhasil diakhiri", lease)
}

func (s *LeaseService) List(c *fiber.Ctx) error {
	status := c.Query("status")
	if status != "" && status != model.LeaseActive && status != model.LeaseEnded {
		return helper.BadRequest("status harus active atau ended")
	}
	leases, err := s.leases.FindAll(c.UserContext(), s.pool, status)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar kontrak berhasil diambil", leases)
}

func (s *LeaseService) Mine(c *fiber.Ctx) error {
	auth := helper.CurrentUser(c)
	leases, err := s.leases.FindByTenant(c.UserContext(), s.pool, auth.UserID)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "kontrak saya berhasil diambil", leases)
}

func (s *LeaseService) Get(c *fiber.Ctx) error {
	id, err := helper.ParseID(c, "id")
	if err != nil {
		return err
	}
	lease, err := s.leases.FindByID(c.UserContext(), s.pool, id)
	if err != nil {
		return lookupError(err, "kontrak")
	}
	auth := helper.CurrentUser(c)
	if auth.Role != model.RoleOwner && lease.TenantID != auth.UserID {
		return helper.NotFound("kontrak tidak ditemukan")
	}
	return helper.Success(c, fiber.StatusOK, "kontrak berhasil diambil", lease)
}
