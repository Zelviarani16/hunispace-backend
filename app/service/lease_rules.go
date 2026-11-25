package service

import (
	"time"

	"hunispace/app/model"
)

const dateLayout = "2006-01-02"

func ValidateLeaseRequest(req model.LeaseCreateRequest) (start, end time.Time, fields map[string]string) {
	fields = map[string]string{}

	if req.TenantID <= 0 {
		fields["tenant_id"] = "wajib diisi"
	}
	if req.RoomID <= 0 {
		fields["room_id"] = "wajib diisi"
	}

	start, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		fields["start_date"] = "format harus YYYY-MM-DD"
	}
	end, err = time.Parse(dateLayout, req.EndDate)
	if err != nil {
		fields["end_date"] = "format harus YYYY-MM-DD"
	}
	if _, bad := fields["start_date"]; !bad {
		if _, bad := fields["end_date"]; !bad && !end.After(start) {
			fields["end_date"] = "harus setelah start_date"
		}
	}
	return start, end, fields
}
