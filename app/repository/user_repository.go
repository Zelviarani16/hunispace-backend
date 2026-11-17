package repository

import (
	"context"

	"hunispace/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, db DBTX, u model.User) (model.User, error)
	FindByID(ctx context.Context, db DBTX, id int64) (model.User, error)
	FindByEmail(ctx context.Context, db DBTX, email string) (model.User, error)
	EnsureOwner(ctx context.Context, db DBTX, u model.User) error
}

type userPostgresRepository struct{}

func NewUserRepository() UserRepository { return &userPostgresRepository{} }

const userColumns = "id, full_name, email, password_hash, phone_number, role, created_at"

func scanUser(row rowScanner) (model.User, error) {
	var u model.User
	err := row.Scan(&u.ID, &u.FullName, &u.Email, &u.PasswordHash, &u.PhoneNumber, &u.Role, &u.CreatedAt)
	return u, err
}

func (r *userPostgresRepository) Create(ctx context.Context, db DBTX, u model.User) (model.User, error) {
	created, err := scanUser(db.QueryRow(ctx,
		`INSERT INTO users (full_name, email, password_hash, phone_number, role)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+userColumns,
		u.FullName, u.Email, u.PasswordHash, u.PhoneNumber, u.Role))
	if err != nil {
		if isUniqueViolation(err) {
			return model.User{}, ErrDuplicate
		}
		return model.User{}, err
	}
	return created, nil
}

func (r *userPostgresRepository) FindByID(ctx context.Context, db DBTX, id int64) (model.User, error) {
	u, err := scanUser(db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
	if err != nil {
		return model.User{}, mapNotFound(err)
	}
	return u, nil
}

func (r *userPostgresRepository) FindByEmail(ctx context.Context, db DBTX, email string) (model.User, error) {
	u, err := scanUser(db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email))
	if err != nil {
		return model.User{}, mapNotFound(err)
	}
	return u, nil
}

func (r *userPostgresRepository) EnsureOwner(ctx context.Context, db DBTX, u model.User) error {
	_, err := db.Exec(ctx,
		`INSERT INTO users (full_name, email, password_hash, phone_number, role)
		 VALUES ($1, $2, $3, $4, 'owner')
		 ON CONFLICT (email) DO NOTHING`,
		u.FullName, u.Email, u.PasswordHash, u.PhoneNumber)
	return err
}
