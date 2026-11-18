package model

import "time"

const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"

	TicketOpen       = "open"
	TicketInProgress = "in_progress"
	TicketResolved   = "resolved"
)

type Ticket struct {
	ID          int64      `json:"id"`
	TenantID    int64      `json:"tenant_id"`
	RoomID      int64      `json:"room_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

type TicketCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

type TicketStatusRequest struct {
	Status  string `json:"status"`
	Version int    `json:"version"`
}
