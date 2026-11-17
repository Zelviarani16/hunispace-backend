package model

import "time"

const (
	LeaseActive = "active"
	LeaseEnded  = "ended"
)

type Lease struct {
	ID        int64     `json:"id"`
	TenantID  int64     `json:"tenant_id"`
	RoomID    int64     `json:"room_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type LeaseCreateRequest struct {
	TenantID  int64  `json:"tenant_id"`
	RoomID    int64  `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
