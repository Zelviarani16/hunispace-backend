package service

import (
	"strings"

	"hunispace/app/model"
)

func validPriority(p string) bool {
	return p == model.PriorityLow || p == model.PriorityMedium || p == model.PriorityHigh
}

func validTicketStatus(s string) bool {
	return s == model.TicketOpen || s == model.TicketInProgress || s == model.TicketResolved
}

func validateTicketCreate(req *model.TicketCreateRequest) map[string]string {
	fields := map[string]string{}
	if t := strings.TrimSpace(req.Title); t == "" || len(t) > 150 {
		fields["title"] = "wajib diisi, maksimal 150 karakter"
	}
	if strings.TrimSpace(req.Description) == "" {
		fields["description"] = "wajib diisi"
	}
	if req.Priority == "" {
		req.Priority = model.PriorityMedium
	}
	if !validPriority(req.Priority) {
		fields["priority"] = "harus low, medium, atau high"
	}
	return fields
}

func validateTicketStatus(req model.TicketStatusRequest) map[string]string {
	fields := map[string]string{}
	if !validTicketStatus(req.Status) {
		fields["status"] = "harus open, in_progress, atau resolved"
	}
	if req.Version < 1 {
		fields["version"] = "wajib diisi (ambil dari data tiket terakhir)"
	}
	return fields
}
