package model

import "time"

const (
	RoomAvailable   = "available"
	RoomOccupied    = "occupied"
	RoomMaintenance = "maintenance"
)

type Room struct {
	ID           int64     `json:"id"`
	RoomNumber   string    `json:"room_number"`
	RoomType     string    `json:"room_type"`
	MonthlyPrice float64   `json:"monthly_price"`
	Facilities   string    `json:"facilities"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type RoomCreateRequest struct {
	RoomNumber   string  `json:"room_number"`
	RoomType     string  `json:"room_type"`
	MonthlyPrice float64 `json:"monthly_price"`
	Facilities   string  `json:"facilities"`
}

type RoomUpdateRequest struct {
	RoomType     *string  `json:"room_type"`
	MonthlyPrice *float64 `json:"monthly_price"`
	Facilities   *string  `json:"facilities"`
	Status       *string  `json:"status"`
}
