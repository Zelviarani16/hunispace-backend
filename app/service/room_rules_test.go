package service

import (
	"testing"

	"hunispace/app/model"
)

func strPtr(s string) *string     { return &s }
func floatPtr(f float64) *float64 { return &f }

func TestValidateRoomCreate(t *testing.T) {
	valid := model.RoomCreateRequest{RoomNumber: "A1", RoomType: "Standar", MonthlyPrice: 1500000}

	tests := []struct {
		name      string
		mutate    func(r *model.RoomCreateRequest)
		wantField string
	}{
		{"data valid", func(r *model.RoomCreateRequest) {}, ""},
		{"nomor kosong", func(r *model.RoomCreateRequest) { r.RoomNumber = "" }, "room_number"},
		{"harga nol", func(r *model.RoomCreateRequest) { r.MonthlyPrice = 0 }, "monthly_price"},
		{"harga negatif", func(r *model.RoomCreateRequest) { r.MonthlyPrice = -1 }, "monthly_price"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			tt.mutate(&req)
			fields := validateRoomCreate(req)
			if tt.wantField == "" && len(fields) != 0 {
				t.Fatalf("seharusnya valid: %v", fields)
			}
			if tt.wantField != "" {
				if _, ok := fields[tt.wantField]; !ok {
					t.Fatalf("seharusnya ada error pada %q: %v", tt.wantField, fields)
				}
			}
		})
	}
}

func TestApplyRoomUpdate(t *testing.T) {
	t.Run("hanya field yang dikirim yang berubah", func(t *testing.T) {
		room := model.Room{RoomType: "Standar", MonthlyPrice: 1000000, Status: model.RoomAvailable}
		fields := applyRoomUpdate(&room, model.RoomUpdateRequest{MonthlyPrice: floatPtr(1200000)})
		if len(fields) != 0 {
			t.Fatalf("seharusnya valid: %v", fields)
		}
		if room.MonthlyPrice != 1200000 || room.RoomType != "Standar" {
			t.Fatalf("hasil tidak sesuai: %+v", room)
		}
	})

	t.Run("status occupied tidak boleh diatur lewat PATCH", func(t *testing.T) {
		room := model.Room{Status: model.RoomAvailable}
		fields := applyRoomUpdate(&room, model.RoomUpdateRequest{Status: strPtr(model.RoomOccupied)})
		if _, ok := fields["status"]; !ok {
			t.Fatal("seharusnya ditolak")
		}
		if room.Status != model.RoomAvailable {
			t.Fatal("room tidak boleh berubah bila ada error")
		}
	})

	t.Run("kamar yang sedang ditempati tidak boleh diubah statusnya", func(t *testing.T) {
		room := model.Room{Status: model.RoomOccupied}
		fields := applyRoomUpdate(&room, model.RoomUpdateRequest{Status: strPtr(model.RoomMaintenance)})
		if _, ok := fields["status"]; !ok {
			t.Fatal("seharusnya ditolak")
		}
	})

	t.Run("available ke maintenance diperbolehkan", func(t *testing.T) {
		room := model.Room{Status: model.RoomAvailable}
		fields := applyRoomUpdate(&room, model.RoomUpdateRequest{Status: strPtr(model.RoomMaintenance)})
		if len(fields) != 0 || room.Status != model.RoomMaintenance {
			t.Fatalf("seharusnya berhasil: %v, %+v", fields, room)
		}
	})
}
