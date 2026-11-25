package service

import (
	"strings"

	"hunispace/app/model"
)

func ValidateRoomCreate(req model.RoomCreateRequest) map[string]string {
	fields := map[string]string{}
	if n := strings.TrimSpace(req.RoomNumber); n == "" || len(n) > 20 {
		fields["room_number"] = "wajib diisi, maksimal 20 karakter"
	}
	if t := strings.TrimSpace(req.RoomType); t == "" || len(t) > 50 {
		fields["room_type"] = "wajib diisi, maksimal 50 karakter"
	}
	if req.MonthlyPrice <= 0 {
		fields["monthly_price"] = "harus lebih dari 0"
	}
	return fields
}

func ApplyRoomUpdate(room *model.Room, req model.RoomUpdateRequest) map[string]string {
	fields := map[string]string{}

	if req.RoomType != nil && (strings.TrimSpace(*req.RoomType) == "" || len(*req.RoomType) > 50) {
		fields["room_type"] = "tidak boleh kosong, maksimal 50 karakter"
	}
	if req.MonthlyPrice != nil && *req.MonthlyPrice <= 0 {
		fields["monthly_price"] = "harus lebih dari 0"
	}
	if req.Status != nil {
		switch {
		case *req.Status != model.RoomAvailable && *req.Status != model.RoomMaintenance:
			fields["status"] = "hanya boleh 'available' atau 'maintenance'"
		case room.Status == model.RoomOccupied:
			fields["status"] = "kamar sedang ditempati; akhiri kontrak terlebih dahulu"
		}
	}
	if len(fields) > 0 {
		return fields
	}

	if req.RoomType != nil {
		room.RoomType = strings.TrimSpace(*req.RoomType)
	}
	if req.MonthlyPrice != nil {
		room.MonthlyPrice = *req.MonthlyPrice
	}
	if req.Facilities != nil {
		room.Facilities = *req.Facilities
	}
	if req.Status != nil {
		room.Status = *req.Status
	}
	return fields
}

func validRoomStatus(s string) bool {
	return s == model.RoomAvailable || s == model.RoomOccupied || s == model.RoomMaintenance
}
