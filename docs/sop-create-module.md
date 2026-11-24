# SOP: Membuat Modul Baru

Prosedur baku menambah modul dengan endpoint CRUD. Modul `rooms` adalah **acuan**: bila ragu, buka berkas padanannya. Contoh di bawah memakai modul fiktif `announcement` (pengumuman).

## Ringkasan langkah

1. Migration
2. Model
3. Repository
4. Aturan murni + test
5. Service
6. Route
7. Build, test, commit

Berkas yang dibuat (semua milik pemilik modul, tidak menyentuh berkas bersama):

```text
migrations/007_announcements.sql
app/model/announcement.go
app/repository/announcement_repository.go
app/service/announcement_rules.go
app/service/announcement_service.go
route/announcement_route.go
tests/announcement_rules_test.go
```

## Langkah 1: Migration

Gunakan `IF NOT EXISTS` agar aman dijalankan berulang, dan `CHECK` untuk nilai enum.

```sql
CREATE TABLE IF NOT EXISTS announcements (
    id         BIGSERIAL PRIMARY KEY,
    title      VARCHAR(150) NOT NULL,
    body       TEXT         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);
```

## Langkah 2: Model

```go
package model

import "time"

type Announcement struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type AnnouncementCreateRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}
```

Request update memakai pointer (`*string`) agar bisa membedakan "tidak dikirim" dari "dikirim kosong".

## Langkah 3: Repository

Kontrak berupa interface; setiap method menerima `DBTX` sehingga dapat dipakai dalam transaction.

```go
type AnnouncementRepository interface {
	Create(ctx context.Context, db DBTX, a model.Announcement) (model.Announcement, error)
	FindAll(ctx context.Context, db DBTX) ([]model.Announcement, error)
}

type announcementPostgresRepository struct{}

func NewAnnouncementRepository() AnnouncementRepository { return &announcementPostgresRepository{} }
```

Gunakan pola yang sudah ada: `scanXxx(row rowScanner)`, `mapNotFound(err)` untuk `pgx.ErrNoRows`, `isUniqueViolation(err)` untuk duplikat, dan `collect(rows, scanXxx)` untuk daftar.

## Langkah 4: Aturan murni + test

Validasi ditulis sebagai fungsi murni yang mengembalikan `map[string]string` (pesan per field; kosong berarti valid), lalu diuji dengan test table-driven di folder `tests/` (package `tests`). Karena test berada di luar package `service`, nama fungsi aturan harus diawali huruf besar (misalnya `ValidateAnnouncement`).

```go
func validateAnnouncement(req model.AnnouncementCreateRequest) map[string]string {
	fields := map[string]string{}
	if strings.TrimSpace(req.Title) == "" || len(req.Title) > 150 {
		fields["title"] = "wajib diisi, maksimal 150 karakter"
	}
	if strings.TrimSpace(req.Body) == "" {
		fields["body"] = "wajib diisi"
	}
	return fields
}
```

Wajib ada baris test di batas (misalnya tepat 150 karakter dan 151 karakter), bukan hanya yang jauh dari batas.

## Langkah 5: Service

Setiap fungsi berbentuk `func (s *XxxService) Nama(c *fiber.Ctx) error`. Kembalikan error dari paket `helper`, jangan menulis response error sendiri.

```go
func (s *AnnouncementService) Create(c *fiber.Ctx) error {
	var req model.AnnouncementCreateRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return err
	}
	if fields := validateAnnouncement(req); len(fields) > 0 {
		return helper.Validation(fields)
	}
	item, err := s.announcements.Create(c.UserContext(), s.pool, model.Announcement{Title: req.Title, Body: req.Body})
	if err != nil {
		return helper.Internal(err)
	}
	return helper.Created(c, "pengumuman berhasil dibuat", item)
}
```

Bila ada beberapa perubahan data yang harus berhasil bersamaan, pakai `TxManager.WithTx` (lihat `LeaseService.Create`).

## Langkah 6: Route

Cukup satu berkas baru. Modul mendaftarkan dirinya lewat `init()`; **jangan mengubah `route.go` atau `main.go`**.

```go
func init() {
	register(func(api fiber.Router, d Deps) {
		svc := service.NewAnnouncementService(d.Pool, repository.NewAnnouncementRepository())

		g := api.Group("/announcements", d.Auth)
		g.Get("", svc.List)
		g.Post("", d.Owner, svc.Create)
	})
}
```

## Langkah 6b: Anotasi Swagger

Tambahkan blok anotasi swaggo TEPAT di atas setiap method service (contoh: `@Summary`, `@Tags`, `@Param`, `@Success`, `@Failure`, `@Router`; lihat `app/service/room_service.go`). Tidak boleh ada baris kosong antara komentar dan `func`. Lalu jalankan `go tool swag init` dan commit hasilnya di folder `docs/`.

## Langkah 7: Build, test, commit

```powershell
gofmt -l .          # harus kosong
go build ./...
go vet ./...
go test ./...
go run .            # AUTO_MIGRATE membuat tabel baru
```

Commit per langkah, misalnya:

```text
feat(announcement): add announcements table migration
feat(announcement): define announcement entity struct
feat(announcement): add announcement repository contract and implementation
feat(announcement): add validation rules
test(announcement): add table-driven tests for validation
feat(announcement): add announcement service
feat(announcement): register announcement routes
```

## Checklist akhir

- [ ] Migration idempotent dan bernomor sesuai pesanan.
- [ ] Hanya berkas milik modul yang berubah (`git status` bersih dari berkas bersama).
- [ ] Endpoint di balik `d.Auth`; role dibatasi `d.Owner` / `d.Tenant` bila perlu.
- [ ] Query milik pengguna difilter pemilik; identitas dari token, bukan body.
- [ ] Test untuk aturan murni, termasuk nilai tepat di batas.
- [ ] `gofmt`, `go build`, `go vet`, `go test` lulus.
- [ ] Tabel endpoint di README diperbarui.
