package model

type DashboardSummary struct {
	RoomsAvailable     int64   `json:"rooms_available"`
	RoomsOccupied      int64   `json:"rooms_occupied"`
	RoomsMaintenance   int64   `json:"rooms_maintenance"`
	ActiveLeases       int64   `json:"active_leases"`
	InvoicesUnpaid     int64   `json:"invoices_unpaid"`
	InvoicesOverdue    int64   `json:"invoices_overdue"`
	InvoicesPending    int64   `json:"invoices_pending_verification"`
	RevenueThisMonth   float64 `json:"revenue_this_month"`
	TicketsNotResolved int64   `json:"tickets_not_resolved"`
}
