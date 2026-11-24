//go:build integration

package integration

import (
	"net/http"
	"sync"
	"testing"
)

func TestAuthAndRoles(t *testing.T) {
	reset(t)
	owner := login(t, ownerEmail, ownerPassword)
	_, tenant := registerTenant(t, "budi@test.local")

	if code, _ := call(t, http.MethodGet, "/api/v1/rooms", "", nil); code != http.StatusUnauthorized {
		t.Errorf("tanpa token: want 401, got %d", code)
	}
	if code, _ := call(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"email": ownerEmail, "password": "salah-salah"}); code != http.StatusUnauthorized {
		t.Errorf("password salah: want 401, got %d", code)
	}
	if code, _ := call(t, http.MethodPost, "/api/v1/auth/register", "", map[string]string{"full_name": "Budi", "email": "budi@test.local", "password": "tenant-pass-123"}); code != http.StatusConflict {
		t.Errorf("email ganda: want 409, got %d", code)
	}
	if code, _ := call(t, http.MethodGet, "/api/v1/users", tenant, nil); code != http.StatusForbidden {
		t.Errorf("tenant ke /users: want 403, got %d", code)
	}
	if code, _ := call(t, http.MethodGet, "/api/v1/users", owner, nil); code != http.StatusOK {
		t.Errorf("owner ke /users: want 200, got %d", code)
	}
}

func TestRoomCRUD(t *testing.T) {
	reset(t)
	owner := login(t, ownerEmail, ownerPassword)
	_, tenant := registerTenant(t, "budi@test.local")

	roomID := createRoom(t, owner, "A1")
	if code, _ := call(t, http.MethodPost, "/api/v1/rooms", owner, map[string]any{"room_number": "A1", "room_type": "Standard", "monthly_price": 1}); code != http.StatusConflict {
		t.Errorf("nomor kamar ganda: want 409, got %d", code)
	}
	if code, _ := call(t, http.MethodPost, "/api/v1/rooms", tenant, map[string]any{"room_number": "B1", "room_type": "Standard", "monthly_price": 1}); code != http.StatusForbidden {
		t.Errorf("tenant buat kamar: want 403, got %d", code)
	}
	path := "/api/v1/rooms/" + itoa(roomID)
	if code, _ := call(t, http.MethodPatch, path, owner, map[string]any{"monthly_price": 1500000}); code != http.StatusOK {
		t.Errorf("update kamar: want 200, got %d", code)
	}
	if code, _ := call(t, http.MethodDelete, path, owner, nil); code != http.StatusOK {
		t.Errorf("hapus kamar: want 200, got %d", code)
	}
	if code, _ := call(t, http.MethodGet, path, owner, nil); code != http.StatusNotFound {
		t.Errorf("kamar terhapus: want 404, got %d", code)
	}
}

// TC double booking: dua owner request bersamaan untuk kamar yang sama,
// hanya satu boleh berhasil (unique index leases_one_active_per_room).
func TestConcurrentDoubleBooking(t *testing.T) {
	reset(t)
	owner := login(t, ownerEmail, ownerPassword)
	roomID := createRoom(t, owner, "A1")
	t1, _ := registerTenant(t, "t1@test.local")
	t2, _ := registerTenant(t, "t2@test.local")

	codes := make([]int, 2)
	var wg sync.WaitGroup
	for i, tenantID := range []int64{t1, t2} {
		wg.Add(1)
		go func(i int, tenantID int64) {
			defer wg.Done()
			codes[i], _ = createLease(t, owner, tenantID, roomID)
		}(i, tenantID)
	}
	wg.Wait()

	ok, conflict := 0, 0
	for _, c := range codes {
		switch c {
		case http.StatusCreated:
			ok++
		case http.StatusConflict:
			conflict++
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("want 1x201 dan 1x409, got %v", codes)
	}
}

// Alur penuh tagihan: buat, tagihan ganda, bayar, bayar ganda, verifikasi.
func TestInvoiceFlowAndDuplicates(t *testing.T) {
	reset(t)
	owner := login(t, ownerEmail, ownerPassword)
	tenantID, tenant := registerTenant(t, "budi@test.local")
	roomID := createRoom(t, owner, "A1")
	code, res := createLease(t, owner, tenantID, roomID)
	if code != http.StatusCreated {
		t.Fatalf("lease: %d %v", code, res)
	}
	leaseID := id(res)

	invoice := map[string]any{"lease_id": leaseID, "billing_period": "2026-02", "amount": 1000000, "due_date": "2026-02-10"}
	code, res = call(t, http.MethodPost, "/api/v1/invoices", owner, invoice)
	if code != http.StatusCreated {
		t.Fatalf("invoice: %d %v", code, res)
	}
	invID := id(res)
	if code, _ := call(t, http.MethodPost, "/api/v1/invoices", owner, invoice); code != http.StatusConflict {
		t.Errorf("tagihan ganda: want 409, got %d", code)
	}

	pay := "/api/v1/invoices/" + itoa(invID) + "/pay"
	body := map[string]string{"payment_method": "transfer", "proof_url": "https://example.com/bukti.jpg"}
	if code, _ := call(t, http.MethodPost, pay, tenant, body); code != http.StatusOK {
		t.Errorf("bayar: want 200, got %d", code)
	}
	if code, _ := call(t, http.MethodPost, pay, tenant, body); code != http.StatusConflict {
		t.Errorf("bayar dua kali: want 409, got %d", code)
	}
	if code, _ := call(t, http.MethodPost, "/api/v1/invoices/"+itoa(invID)+"/verify", owner, nil); code != http.StatusOK {
		t.Errorf("verifikasi: want 200, got %d", code)
	}
	if code, _ := call(t, http.MethodPost, "/api/v1/invoices/"+itoa(invID)+"/verify", owner, nil); code == http.StatusOK {
		t.Errorf("verifikasi kedua tidak boleh 200")
	}
}

// Pembuatan tagihan yang sama secara bersamaan: hanya satu yang tercipta.
func TestConcurrentDuplicateInvoice(t *testing.T) {
	reset(t)
	owner := login(t, ownerEmail, ownerPassword)
	tenantID, _ := registerTenant(t, "budi@test.local")
	roomID := createRoom(t, owner, "A1")
	_, res := createLease(t, owner, tenantID, roomID)
	leaseID := id(res)

	const n = 5
	codes := make([]int, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			codes[i], _ = call(t, http.MethodPost, "/api/v1/invoices", owner, map[string]any{
				"lease_id": leaseID, "billing_period": "2026-03", "amount": 1000000, "due_date": "2026-03-10"})
		}(i)
	}
	wg.Wait()

	created := 0
	for _, c := range codes {
		if c == http.StatusCreated {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("want tepat 1x201, got %v", codes)
	}
}

func TestMaintenanceAndDashboard(t *testing.T) {
	reset(t)
	owner := login(t, ownerEmail, ownerPassword)
	tenantID, tenant := registerTenant(t, "budi@test.local")
	roomID := createRoom(t, owner, "A1")
	if code, _ := createLease(t, owner, tenantID, roomID); code != http.StatusCreated {
		t.Fatalf("lease gagal: %d", code)
	}

	code, res := call(t, http.MethodPost, "/api/v1/tickets", tenant, map[string]any{
		"title": "AC bocor", "description": "Menetes ke lantai", "priority": "high"})
	if code != http.StatusCreated {
		t.Fatalf("tiket: %d %v", code, res)
	}
	path := "/api/v1/tickets/" + itoa(id(res)) + "/status"
	version := int(data(res)["version"].(float64))
	if code, _ := call(t, http.MethodPatch, path, owner, map[string]any{"status": "in_progress", "version": version}); code != http.StatusOK {
		t.Errorf("ubah status: want 200, got %d", code)
	}
	// version lama dipakai lagi: ditolak (optimistic locking)
	if code, _ := call(t, http.MethodPatch, path, owner, map[string]any{"status": "resolved", "version": version}); code != http.StatusConflict {
		t.Errorf("version basi: want 409, got %d", code)
	}
	if code, _ := call(t, http.MethodGet, "/api/v1/dashboard/summary", owner, nil); code != http.StatusOK {
		t.Errorf("dashboard: want 200, got %d", code)
	}
	if code, _ := call(t, http.MethodGet, "/api/v1/dashboard/summary", tenant, nil); code != http.StatusForbidden {
		t.Errorf("dashboard tenant: want 403, got %d", code)
	}
}
