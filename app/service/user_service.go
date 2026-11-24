package service

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"hunispace/app/model"
	"hunispace/app/repository"
	"hunispace/helper"
)

type UserService struct {
	pool  *pgxpool.Pool
	users repository.UserRepository
	jwt   *helper.JWTManager
}

func NewUserService(pool *pgxpool.Pool, users repository.UserRepository, jwt *helper.JWTManager) *UserService {
	return &UserService{pool: pool, users: users, jwt: jwt}
}

// Register: POST /api/v1/auth/register. Akun baru SELALU berperan tenant.
func (s *UserService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FullName = strings.TrimSpace(req.FullName)

	if fields := ValidateRegister(req); len(fields) > 0 {
		return helper.Validation(fields)
	}

	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Internal(err)
	}

	user, err := s.users.Create(c.UserContext(), s.pool, model.User{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hash,
		PhoneNumber:  req.PhoneNumber,
		Role:         model.RoleTenant,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Conflict("EMAIL_TAKEN", "email sudah terdaftar")
		}
		return helper.Internal(err)
	}
	return helper.Created(c, "registrasi berhasil", user)
}

func (s *UserService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || req.Password == "" {
		return helper.Validation(map[string]string{"email": "wajib diisi", "password": "wajib diisi"})
	}

	user, err := s.users.FindByEmail(c.UserContext(), s.pool, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Unauthorized("email atau password salah")
		}
		return helper.Internal(err)
	}
	if !helper.VerifyPassword(user.PasswordHash, req.Password) {
		return helper.Unauthorized("email atau password salah")
	}

	token, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "login berhasil", model.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwt.AccessTTL().Seconds()),
		User:        user,
	})
}

func (s *UserService) Me(c *fiber.Ctx) error {
	auth := helper.CurrentUser(c)
	user, err := s.users.FindByID(c.UserContext(), s.pool, auth.UserID)
	if err != nil {
		return lookupError(err, "pengguna")
	}
	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", user)
}

func (s *UserService) EnsureOwner(ctx context.Context, name, email, password string) error {
	hash, err := helper.HashPassword(password)
	if err != nil {
		return err
	}
	return s.users.EnsureOwner(ctx, s.pool, model.User{
		FullName:     name,
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: hash,
	})
}

func (s *UserService) UpdateMe(c *fiber.Ctx) error {
	var req model.UpdateProfileRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	if fields := ValidateProfileUpdate(req); len(fields) > 0 {
		return helper.Validation(fields)
	}
	ctx := c.UserContext()
	auth := helper.CurrentUser(c)
	current, err := s.users.FindByID(ctx, s.pool, auth.UserID)
	if err != nil {
		return lookupError(err, "pengguna")
	}
	name, phone := current.FullName, current.PhoneNumber
	if req.FullName != nil {
		name = strings.TrimSpace(*req.FullName)
	}
	if req.PhoneNumber != nil {
		phone = *req.PhoneNumber
	}
	user, err := s.users.UpdateProfile(ctx, s.pool, auth.UserID, name, phone)
	if err != nil {
		return lookupError(err, "pengguna")
	}
	return helper.Success(c, fiber.StatusOK, "profil berhasil diperbarui", user)
}

func (s *UserService) ChangePassword(c *fiber.Ctx) error {
	var req model.ChangePasswordRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	if fields := ValidateChangePassword(req); len(fields) > 0 {
		return helper.Validation(fields)
	}
	ctx := c.UserContext()
	auth := helper.CurrentUser(c)
	user, err := s.users.FindByID(ctx, s.pool, auth.UserID)
	if err != nil {
		return lookupError(err, "pengguna")
	}
	if !helper.VerifyPassword(user.PasswordHash, req.OldPassword) {
		return helper.Unauthorized("password lama salah")
	}
	hash, err := helper.HashPassword(req.NewPassword)
	if err != nil {
		return helper.Internal(err)
	}
	if err := s.users.UpdatePassword(ctx, s.pool, auth.UserID, hash); err != nil {
		return lookupError(err, "pengguna")
	}
	return helper.Success(c, fiber.StatusOK, "password berhasil diubah", nil)
}

func (s *UserService) List(c *fiber.Ctx) error {
	role := c.Query("role")
	if role != "" && role != model.RoleOwner && role != model.RoleTenant {
		return helper.BadRequest("role harus owner atau tenant")
	}
	users, err := s.users.FindAll(c.UserContext(), s.pool, role)
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pengguna berhasil diambil", users)
}
