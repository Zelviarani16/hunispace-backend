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

type RoomService struct {
	pool  *pgxpool.Pool
	rooms repository.RoomRepository
}

func NewRoomService(pool *pgxpool.Pool, rooms repository.RoomRepository) *RoomService {
	return &RoomService{pool: pool, rooms: rooms}
}

func (s *RoomService) List(c *fiber.Ctx) error {
	status := c.Query("status")
	if status != "" && !validRoomStatus(status) {
		return helper.BadRequest("status harus available, occupied, atau maintenance")
	}
	rooms, err := s.rooms.FindAll(c.UserContext(), s.pool, status)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar kamar berhasil diambil", rooms)
}

func (s *RoomService) Get(c *fiber.Ctx) error {
	id, err := helper.ParseID(c, "id")
	if err != nil {
		return err
	}
	room, err := s.rooms.FindByID(c.UserContext(), s.pool, id)
	if err != nil {
		return lookupError(err, "kamar")
	}
	return helper.Success(c, fiber.StatusOK, "kamar berhasil diambil", room)
}

func (s *RoomService) Create(c *fiber.Ctx) error {
	var req model.RoomCreateRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	if fields := validateRoomCreate(req); len(fields) > 0 {
		return helper.Validation(fields)
	}

	room, err := s.rooms.Create(c.UserContext(), s.pool, model.Room{
		RoomNumber:   strings.TrimSpace(req.RoomNumber),
		RoomType:     strings.TrimSpace(req.RoomType),
		MonthlyPrice: req.MonthlyPrice,
		Facilities:   req.Facilities,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("ROOM_NUMBER_TAKEN", "nomor kamar sudah dipakai")
		}
		return helper.Internal(err)
	}
	return helper.Created(c, "kamar berhasil dibuat", room)
}

func (s *RoomService) Update(c *fiber.Ctx) error {
	id, err := helper.ParseID(c, "id")
	if err != nil {
		return err
	}
	var req model.RoomUpdateRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}

	ctx := c.UserContext()
	room, err := s.rooms.FindByID(ctx, s.pool, id)
	if err != nil {
		return lookupError(err, "kamar")
	}
	if fields := applyRoomUpdate(&room, req); len(fields) > 0 {
		return helper.Validation(fields)
	}

	updated, err := s.rooms.Update(ctx, s.pool, room)
	if err != nil {
		return lookupError(err, "kamar")
	}
	return helper.Success(c, fiber.StatusOK, "kamar berhasil diperbarui", updated)
}

func (s *RoomService) Delete(c *fiber.Ctx) error {
	id, err := helper.ParseID(c, "id")
	if err != nil {
		return err
	}
	if err := s.rooms.Delete(c.UserContext(), s.pool, id); err != nil {
		if errors.Is(err, repository.ErrInUse) {
			return helper.Conflict("ROOM_IN_USE", "kamar masih punya kontrak sewa sehingga tidak dapat dihapus")
		}
		return lookupError(err, "kamar")
	}
	return helper.Success(c, fiber.StatusOK, "kamar berhasil dihapus", nil)
}
