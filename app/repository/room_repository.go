package repository

import (
	"context"

	"hunispace/app/model"
)

type RoomRepository interface {
	Create(ctx context.Context, db DBTX, r model.Room) (model.Room, error)
	FindAll(ctx context.Context, db DBTX, status string) ([]model.Room, error)
	FindByID(ctx context.Context, db DBTX, id int64) (model.Room, error)
	FindByIDForUpdate(ctx context.Context, db DBTX, id int64) (model.Room, error)
	Update(ctx context.Context, db DBTX, r model.Room) (model.Room, error)
	UpdateStatus(ctx context.Context, db DBTX, id int64, status string) error
	Delete(ctx context.Context, db DBTX, id int64) error
}

type roomPostgresRepository struct{}

func NewRoomRepository() RoomRepository { return &roomPostgresRepository{} }

const roomColumns = "id, room_number, room_type, monthly_price, facilities, status, created_at"

func scanRoom(row rowScanner) (model.Room, error) {
	var r model.Room
	err := row.Scan(&r.ID, &r.RoomNumber, &r.RoomType, &r.MonthlyPrice, &r.Facilities, &r.Status, &r.CreatedAt)
	return r, err
}

func (r *roomPostgresRepository) Create(ctx context.Context, db DBTX, room model.Room) (model.Room, error) {
	created, err := scanRoom(db.QueryRow(ctx,
		`INSERT INTO rooms (room_number, room_type, monthly_price, facilities)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+roomColumns,
		room.RoomNumber, room.RoomType, room.MonthlyPrice, room.Facilities))
	if err != nil {
		if isUniqueViolation(err) {
			return model.Room{}, ErrDuplicate
		}
		return model.Room{}, err
	}
	return created, nil
}

func (r *roomPostgresRepository) FindAll(ctx context.Context, db DBTX, status string) ([]model.Room, error) {
	query := `SELECT ` + roomColumns + ` FROM rooms`
	args := []any{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY room_number`

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanRoom)
}

func (r *roomPostgresRepository) FindByID(ctx context.Context, db DBTX, id int64) (model.Room, error) {
	room, err := scanRoom(db.QueryRow(ctx, `SELECT `+roomColumns+` FROM rooms WHERE id = $1`, id))
	if err != nil {
		return model.Room{}, mapNotFound(err)
	}
	return room, nil
}

func (r *roomPostgresRepository) FindByIDForUpdate(ctx context.Context, db DBTX, id int64) (model.Room, error) {
	room, err := scanRoom(db.QueryRow(ctx,
		`SELECT `+roomColumns+` FROM rooms WHERE id = $1 FOR UPDATE`, id))
	if err != nil {
		return model.Room{}, mapNotFound(err)
	}
	return room, nil
}

func (r *roomPostgresRepository) Update(ctx context.Context, db DBTX, room model.Room) (model.Room, error) {
	updated, err := scanRoom(db.QueryRow(ctx,
		`UPDATE rooms SET room_type = $2, monthly_price = $3, facilities = $4, status = $5
		 WHERE id = $1
		 RETURNING `+roomColumns,
		room.ID, room.RoomType, room.MonthlyPrice, room.Facilities, room.Status))
	if err != nil {
		return model.Room{}, mapNotFound(err)
	}
	return updated, nil
}

func (r *roomPostgresRepository) UpdateStatus(ctx context.Context, db DBTX, id int64, status string) error {
	tag, err := db.Exec(ctx, `UPDATE rooms SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *roomPostgresRepository) Delete(ctx context.Context, db DBTX, id int64) error {
	tag, err := db.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, id)
	if err != nil {
		if isForeignKeyViolation(err) {
			return ErrInUse
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
