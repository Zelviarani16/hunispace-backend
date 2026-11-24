//go:build integration

// Package integration menguji HuniSpace dari HTTP sampai PostgreSQL asli.
//
// Cara menjalankan (database HARUS berakhiran _test, karena tabelnya dikosongkan):
//
//	$env:TEST_DATABASE_URL="postgres://postgres:PASSWORD@localhost:5432/hunispace_test?sslmode=disable"
//	go test -tags integration -race -shuffle=on ./tests/integration/...
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"hunispace/app/model"
	"hunispace/app/repository"
	"hunispace/app/service"
	"hunispace/database"
	"hunispace/helper"
	"hunispace/middleware"
	"hunispace/route"
)

const (
	ownerEmail    = "owner@test.local"
	ownerPassword = "owner-pass-123"
)

var (
	pool *pgxpool.Pool
	app  *fiber.App
)

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Println("TEST_DATABASE_URL kosong: integration test dilewati")
		os.Exit(0)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fmt.Println("TEST_DATABASE_URL tidak valid:", err)
		os.Exit(1)
	}
	if !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		fmt.Printf("nama database %q harus berakhiran _test\n", cfg.ConnConfig.Database)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		fmt.Println("koneksi database gagal:", err)
		os.Exit(1)
	}
	if err := database.Migrate(ctx, pool, ""); err != nil {
		fmt.Println("migration gagal:", err)
		os.Exit(1)
	}

	jwt := helper.NewJWTManager("integration-test-secret-minimal-32-char!!", "hunispace-test", time.Hour)
	app = fiber.New()
	app.Use(middleware.ErrorResponse())
	route.Setup(app, route.Deps{
		Pool:   pool,
		Tx:     repository.NewTxManager(pool),
		JWT:    jwt,
		Auth:   middleware.RequireAuth(jwt),
		Owner:  middleware.RequireRole(model.RoleOwner),
		Tenant: middleware.RequireRole(model.RoleTenant),
	})

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

// reset mengosongkan semua tabel dan membuat ulang akun owner.
func reset(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `TRUNCATE maintenance_tickets, invoices, leases, rooms, users RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	jwt := helper.NewJWTManager("integration-test-secret-minimal-32-char!!", "hunispace-test", time.Hour)
	users := service.NewUserService(pool, repository.NewUserRepository(), jwt)
	if err := users.EnsureOwner(ctx, "Owner", ownerEmail, ownerPassword); err != nil {
		t.Fatalf("seed owner: %v", err)
	}
}

// call mengirim request ke app tanpa membuka port, lalu mengurai JSON balasan.
func call(t *testing.T, method, path, token string, body any) (int, map[string]any) {
	t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return res.StatusCode, out
}

func data(m map[string]any) map[string]any {
	d, _ := m["data"].(map[string]any)
	return d
}

func id(m map[string]any) int64 {
	return int64(data(m)["id"].(float64))
}

func login(t *testing.T, email, password string) string {
	t.Helper()
	code, res := call(t, http.MethodPost, "/api/v1/auth/login", "", map[string]string{"email": email, "password": password})
	if code != http.StatusOK {
		t.Fatalf("login %s: status %d %v", email, code, res)
	}
	return data(res)["access_token"].(string)
}

// registerTenant mendaftarkan penyewa baru dan mengembalikan id dan token-nya.
func registerTenant(t *testing.T, email string) (int64, string) {
	t.Helper()
	code, res := call(t, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"full_name": "Penyewa " + email, "email": email, "password": "tenant-pass-123",
	})
	if code != http.StatusCreated {
		t.Fatalf("register %s: status %d %v", email, code, res)
	}
	return id(res), login(t, email, "tenant-pass-123")
}

func createRoom(t *testing.T, owner, number string) int64 {
	t.Helper()
	code, res := call(t, http.MethodPost, "/api/v1/rooms", owner, map[string]any{
		"room_number": number, "room_type": "Standard", "monthly_price": 1000000, "facilities": "AC",
	})
	if code != http.StatusCreated {
		t.Fatalf("create room: status %d %v", code, res)
	}
	return id(res)
}

func createLease(t *testing.T, owner string, tenantID, roomID int64) (int, map[string]any) {
	t.Helper()
	return call(t, http.MethodPost, "/api/v1/leases", owner, map[string]any{
		"tenant_id": tenantID, "room_id": roomID, "start_date": "2026-01-01", "end_date": "2026-12-31",
	})
}

func itoa(n int64) string { return fmt.Sprintf("%d", n) }
