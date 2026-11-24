package repository

import (
	"context"

	"hunispace/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, db DBTX, u model.User) (model.User, error)
	FindByID(ctx context.Context, db DBTX, id int64) (model.User, error)
	FindByEmail(ctx context.Context, db DBTX, email string) (model.User, error)
	// EnsureOwner membuat akun owner bila emailnya belum ada.
	EnsureOwner(ctx context.Context, db DBTX, u model.User) error
	FindAll(ctx context.Context, db DBTX, role string) ([]model.User, error)
	UpdateProfile(ctx context.Context, db DBTX, id int64, fullName, phone string) (model.User, error)
	UpdatePassword(ctx context.Context, db DBTX, id int64, hash string) error
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

func (r *userPostgresRepository) FindAll(ctx context.Context, db DBTX, role string) ([]model.User, error) {
	query := `SELECT ` + userColumns + ` FROM users`
	args := []any{}
	if role != "" {
		query += ` WHERE role = $1`
		args = append(args, role)
	}
	query += ` ORDER BY id`
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanUser)
}

func (r *userPostgresRepository) UpdateProfile(ctx context.Context, db DBTX, id int64, fullName, phone string) (model.User, error) {
	u, err := scanUser(db.QueryRow(ctx,
		`UPDATE users SET full_name = $2, phone_number = $3 WHERE id = $1
		 RETURNING `+userColumns, id, fullName, phone))
	if err != nil {
		return model.User{}, mapNotFound(err)
	}
	return u, nil
}

func (r *userPostgresRepository) UpdatePassword(ctx context.Context, db DBTX, id int64, hash string) error {
	tag, err := db.Exec(ctx, `UPDATE users SET password_hash = $2 WHERE id = $1`, id, hash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
