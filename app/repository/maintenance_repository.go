package repository

import (
	"context"

	"hunispace/app/model"
)

type MaintenanceRepository interface {
	Create(ctx context.Context, db DBTX, t model.Ticket) (model.Ticket, error)
	FindByID(ctx context.Context, db DBTX, id int64) (model.Ticket, error)
	FindAll(ctx context.Context, db DBTX, status string) ([]model.Ticket, error)
	FindByTenant(ctx context.Context, db DBTX, tenantID int64) ([]model.Ticket, error)
	UpdateStatus(ctx context.Context, db DBTX, id int64, status string, version int) (model.Ticket, error)
}

type maintenancePostgresRepository struct{}

func NewMaintenanceRepository() MaintenanceRepository { return &maintenancePostgresRepository{} }

const ticketColumns = "id, tenant_id, room_id, title, description, priority, status, version, created_at, updated_at, resolved_at"

func scanTicket(row rowScanner) (model.Ticket, error) {
	var t model.Ticket
	err := row.Scan(&t.ID, &t.TenantID, &t.RoomID, &t.Title, &t.Description, &t.Priority,
		&t.Status, &t.Version, &t.CreatedAt, &t.UpdatedAt, &t.ResolvedAt)
	return t, err
}

func (r *maintenancePostgresRepository) Create(ctx context.Context, db DBTX, t model.Ticket) (model.Ticket, error) {
	return scanTicket(db.QueryRow(ctx,
		`INSERT INTO maintenance_tickets (tenant_id, room_id, title, description, priority)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+ticketColumns,
		t.TenantID, t.RoomID, t.Title, t.Description, t.Priority))
}

func (r *maintenancePostgresRepository) FindByID(ctx context.Context, db DBTX, id int64) (model.Ticket, error) {
	t, err := scanTicket(db.QueryRow(ctx,
		`SELECT `+ticketColumns+` FROM maintenance_tickets WHERE id = $1`, id))
	if err != nil {
		return model.Ticket{}, mapNotFound(err)
	}
	return t, nil
}

func (r *maintenancePostgresRepository) FindAll(ctx context.Context, db DBTX, status string) ([]model.Ticket, error) {
	query := `SELECT ` + ticketColumns + ` FROM maintenance_tickets`
	args := []any{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanTicket)
}

func (r *maintenancePostgresRepository) FindByTenant(ctx context.Context, db DBTX, tenantID int64) ([]model.Ticket, error) {
	rows, err := db.Query(ctx,
		`SELECT `+ticketColumns+` FROM maintenance_tickets
		 WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanTicket)
}

func (r *maintenancePostgresRepository) UpdateStatus(
	ctx context.Context, db DBTX, id int64, status string, version int,
) (model.Ticket, error) {
	t, err := scanTicket(db.QueryRow(ctx,
		`UPDATE maintenance_tickets
		 SET status = $2::varchar,
		     version = version + 1,
		     updated_at = now(),
		     resolved_at = CASE WHEN $2::varchar = 'resolved' THEN now() ELSE NULL END
		 WHERE id = $1 AND version = $3
		 RETURNING `+ticketColumns,
		id, status, version))
	if err != nil {
		return model.Ticket{}, mapNotFound(err)
	}
	return t, nil
}
