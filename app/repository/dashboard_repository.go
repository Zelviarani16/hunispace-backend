package repository

import (
	"context"

	"hunispace/app/model"
)

type DashboardRepository interface {
	Summary(ctx context.Context, db DBTX) (model.DashboardSummary, error)
}

type dashboardPostgresRepository struct{}

func NewDashboardRepository() DashboardRepository { return &dashboardPostgresRepository{} }

func (r *dashboardPostgresRepository) Summary(ctx context.Context, db DBTX) (model.DashboardSummary, error) {
	var s model.DashboardSummary
	err := db.QueryRow(ctx, `
		SELECT
		  (SELECT count(*) FROM rooms WHERE status = 'available'),
		  (SELECT count(*) FROM rooms WHERE status = 'occupied'),
		  (SELECT count(*) FROM rooms WHERE status = 'maintenance'),
		  (SELECT count(*) FROM leases WHERE status = 'active'),
		  (SELECT count(*) FROM invoices WHERE status = 'unpaid'),
		  (SELECT count(*) FROM invoices WHERE status = 'overdue'),
		  (SELECT count(*) FROM invoices WHERE status = 'pending_verification'),
		  (SELECT COALESCE(SUM(amount), 0)::float8 FROM invoices
		     WHERE status = 'paid' AND date_trunc('month', paid_at) = date_trunc('month', now())),
		  (SELECT count(*) FROM maintenance_tickets WHERE status <> 'resolved')`).
		Scan(&s.RoomsAvailable, &s.RoomsOccupied, &s.RoomsMaintenance, &s.ActiveLeases,
			&s.InvoicesUnpaid, &s.InvoicesOverdue, &s.InvoicesPending,
			&s.RevenueThisMonth, &s.TicketsNotResolved)
	return s, err
}
