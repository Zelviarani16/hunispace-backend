package repository

import (
	"context"

	"hunispace/app/model"
)

type LeaseRepository interface {
	Create(ctx context.Context, db DBTX, l model.Lease) (model.Lease, error)
	FindByID(ctx context.Context, db DBTX, id int64) (model.Lease, error)
	FindByIDForUpdate(ctx context.Context, db DBTX, id int64) (model.Lease, error)
	FindAll(ctx context.Context, db DBTX, status string) ([]model.Lease, error)
	FindByTenant(ctx context.Context, db DBTX, tenantID int64) ([]model.Lease, error)
	FindActiveByTenant(ctx context.Context, db DBTX, tenantID int64) (model.Lease, error)
	SetStatus(ctx context.Context, db DBTX, id int64, status string) error
}

type leasePostgresRepository struct{}

func NewLeaseRepository() LeaseRepository { return &leasePostgresRepository{} }

const leaseColumns = "id, tenant_id, room_id, start_date, end_date, status, created_at"

func scanLease(row rowScanner) (model.Lease, error) {
	var l model.Lease
	err := row.Scan(&l.ID, &l.TenantID, &l.RoomID, &l.StartDate, &l.EndDate, &l.Status, &l.CreatedAt)
	return l, err
}

func (r *leasePostgresRepository) Create(ctx context.Context, db DBTX, l model.Lease) (model.Lease, error) {
	created, err := scanLease(db.QueryRow(ctx,
		`INSERT INTO leases (tenant_id, room_id, start_date, end_date)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+leaseColumns,
		l.TenantID, l.RoomID, l.StartDate, l.EndDate))
	if err != nil {
		if isUniqueViolation(err) {
			return model.Lease{}, ErrDuplicate
		}
		return model.Lease{}, err
	}
	return created, nil
}

func (r *leasePostgresRepository) FindByID(ctx context.Context, db DBTX, id int64) (model.Lease, error) {
	l, err := scanLease(db.QueryRow(ctx, `SELECT `+leaseColumns+` FROM leases WHERE id = $1`, id))
	if err != nil {
		return model.Lease{}, mapNotFound(err)
	}
	return l, nil
}

func (r *leasePostgresRepository) FindByIDForUpdate(ctx context.Context, db DBTX, id int64) (model.Lease, error) {
	l, err := scanLease(db.QueryRow(ctx,
		`SELECT `+leaseColumns+` FROM leases WHERE id = $1 FOR UPDATE`, id))
	if err != nil {
		return model.Lease{}, mapNotFound(err)
	}
	return l, nil
}

func (r *leasePostgresRepository) FindAll(ctx context.Context, db DBTX, status string) ([]model.Lease, error) {
	query := `SELECT ` + leaseColumns + ` FROM leases`
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
	return collect(rows, scanLease)
}

func (r *leasePostgresRepository) FindByTenant(ctx context.Context, db DBTX, tenantID int64) ([]model.Lease, error) {
	rows, err := db.Query(ctx,
		`SELECT `+leaseColumns+` FROM leases WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanLease)
}

func (r *leasePostgresRepository) FindActiveByTenant(ctx context.Context, db DBTX, tenantID int64) (model.Lease, error) {
	l, err := scanLease(db.QueryRow(ctx,
		`SELECT `+leaseColumns+` FROM leases
		 WHERE tenant_id = $1 AND status = 'active'
		 ORDER BY created_at DESC LIMIT 1`, tenantID))
	if err != nil {
		return model.Lease{}, mapNotFound(err)
	}
	return l, nil
}

func (r *leasePostgresRepository) SetStatus(ctx context.Context, db DBTX, id int64, status string) error {
	tag, err := db.Exec(ctx, `UPDATE leases SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
