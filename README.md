# HuniSpace: Smart Boarding House & Tenant Management Backend

> Go • Fiber v2 • PostgreSQL • JWT • Clean Architecture • Docker • Swagger

**HuniSpace** adalah sistem backend REST API yang mendigitalisasi operasional rumah kos secara terpusat: ketersediaan kamar, kontrak sewa, tagihan dan pembayaran bulanan, serta laporan keluhan fasilitas. Sistem menghubungkan **pemilik kos (owner)** dengan **penghuni (tenant)**.

Repository: `https://github.com/Zelviarani16/hunispace-backend`

---

## Tim

| Anggota | Peran | Modul |
|---|---|---|
| Zelvia Rani F.D.S. | **Git Master**, fondasi, review & merge | Config, database, helper, middleware, route, `users` |
| Drefita Putri C. | Developer | `rooms` dan `leases` |
| Putri Apriliana | Developer | `invoices` dan `maintenance_tickets` |

---

## Latar Belakang

Pengelolaan rumah kos umumnya masih manual: ketersediaan kamar dicatat di buku atau pesan singkat, tagihan bulanan ditagih satu per satu, bukti pembayaran dikirim lewat chat, dan keluhan kerusakan fasilitas mudah terlewat. Akibatnya data tidak terpusat, kamar dapat tersewa ganda, tagihan dapat tercatat dua kali, dan progres perbaikan sulit dipantau oleh penghuni maupun pemilik.

HuniSpace mendigitalisasi proses tersebut lewat REST API yang dapat diintegrasikan ke aplikasi web maupun mobile.

---

## Role dan Hak Akses

| Role | Hak akses |
|---|---|
| **owner** | Mengelola kamar (CRUD), membuat dan mengakhiri kontrak sewa, menerbitkan tagihan, memverifikasi atau menolak pembayaran, memproses tiket keluhan, melihat daftar pengguna dan ringkasan dashboard |
| **tenant** | Melihat kamar, melihat kontrak dan tagihan miliknya, mengonfirmasi pembayaran, membuat dan memantau tiket keluhan |

> Registrasi publik selalu menghasilkan akun **tenant**. Akun **owner** pertama dibuat lewat konfigurasi (`SEED_OWNER_EMAIL` dan `SEED_OWNER_PASSWORD`), bukan lewat endpoint, agar tidak ada yang dapat mendaftarkan diri sebagai owner.

---

## Fitur

| No | Modul | Fitur | Role |
|---|---|---|---|
| 1 | Auth & RBAC | Registrasi, login JWT, profil, ubah nama dan telepon, ganti password (butuh password lama), daftar pengguna, pembatasan endpoint berdasarkan role | Semua (daftar pengguna: Owner) |
| 2 | Kamar | Tambah, lihat (filter status), ubah sebagian, hapus; status `available` / `occupied` / `maintenance` | Owner (kelola), Tenant (lihat) |
| 3 | Kontrak sewa | Menetapkan penghuni ke kamar dengan rentang tanggal; status kamar berubah otomatis; mengakhiri kontrak; detail kontrak | Owner, Tenant |
| 4 | Tagihan | Menerbitkan tagihan per periode; penghuni mengonfirmasi pembayaran dengan bukti; owner memverifikasi atau menolak bukti dengan alasan; penandaan tagihan lewat jatuh tempo | Owner, Tenant |
| 5 | Tiket keluhan | Penghuni melapor kerusakan fasilitas; owner mengubah status `open` → `in_progress` → `resolved` | Tenant (lapor), Owner (proses) |
| 6 | Dashboard | Ringkasan kamar, kontrak, tagihan, pendapatan bulan ini, dan tiket | Owner |
| 7 | Dokumentasi API | Swagger UI (swaggo, code-first), aktif lewat `SWAGGER_ENABLED=true` | Semua |

---

## Arsitektur

Proyek menerapkan **Clean Architecture** (mengikuti `latihan-fiber` di modul perkuliahan). **Tidak ada lapisan controller maupun handler**: fungsi penanganan HTTP berada di dalam `service`, sedangkan `route` hanya mendaftarkan path dan middleware.

```
route → service → repository → database
           │            │
           └─── model ──┘
```

Aturan arah ketergantungan: **lapisan atas memanggil lapisan bawah, tidak pernah sebaliknya.**
`route → service → repository → model`

| Lapisan | Istilah umum | Folder | Tugas | Tidak boleh |
|---|---|---|---|---|
| Entity | entity | `app/model` | Struct domain, request, konstanta | Logika bisnis |
| Repository | repository | `app/repository` | Interface (kontrak) + implementasi query pgx | Aturan bisnis, mengenal Fiber |
| Service | usecase + delivery | `app/service` | Validasi, aturan bisnis, membaca input, membungkus response, fungsi penanganan Fiber | Menulis SQL langsung |
| Route | routing | `route` | Mendaftarkan path dan middleware per modul | Logika bisnis |

### Struktur Folder

```
hunispace/
├── main.go                 # bootstrap server, migration, graceful shutdown, info umum Swagger
├── config/                 # pemuat environment (.env via godotenv)
├── database/               # koneksi pgxpool dan migration runner
├── helper/                 # AppError, response, JWT, bcrypt, parser request
├── middleware/             # RequireAuth, RequireRole, ErrorResponse
├── migrations/             # 001_users … 005_maintenance_tickets, 006_invoice_rejection
├── route/                  # route.go + satu berkas route per modul
├── app/
│   ├── model/              # user, room, lease, invoice, maintenance, common
│   ├── repository/         # kontrak (interface) dan implementasi per modul
│   └── service/            # service, *_rules.go (aturan murni, fungsi diekspor), anotasi Swagger
├── tests/                  # unit test aturan murni (package tests), terpisah dari kode
├── Dockerfile              # image aplikasi
├── docker-compose.yml      # container aplikasi + PostgreSQL
└── docs/                   # workflow, SOP modul, skema database, dokumentasi API, hasil swag init
```

> **Pendaftaran route otomatis.** Setiap `*_route.go` mendaftarkan dirinya lewat `init()` ke `route.register`. Menambah modul tidak perlu mengubah `main.go` maupun `route.go`, anggota tim tidak saling bentrok di berkas bersama.

---

## Rancangan Database

Lima tabel: dua master (`users`, `rooms`) dan tiga transaksi (`leases`, `invoices`, `maintenance_tickets`). Semua migration idempotent (`IF NOT EXISTS`) dan dijalankan otomatis saat server menyala. Rincian kolom dan constraint ada di [docs/database-schema.md](docs/database-schema.md).

```mermaid
erDiagram
    users ||--o{ leases : "menyewa"
    rooms ||--o{ leases : "disewa"
    leases ||--o{ invoices : "ditagih"
    users ||--o{ maintenance_tickets : "melapor"
    rooms ||--o{ maintenance_tickets : "dilaporkan"

    users {
        bigint id PK
        string full_name
        string email UK
        string password_hash
        string phone_number
        string role "owner | tenant"
        timestamptz created_at
    }
    rooms {
        bigint id PK
        string room_number UK
        string room_type
        numeric monthly_price
        text facilities
        string status "available | occupied | maintenance"
        timestamptz created_at
    }
    leases {
        bigint id PK
        bigint tenant_id FK
        bigint room_id FK
        date start_date
        date end_date
        string status "active | ended"
        timestamptz created_at
    }
    invoices {
        bigint id PK
        bigint lease_id FK
        string invoice_number UK
        string billing_period "YYYY-MM, UK+lease_id"
        numeric amount
        date due_date
        string status "unpaid | pending_verification | paid | overdue"
        text rejection_note "alasan penolakan bukti bayar (migrasi 006)"
        timestamptz paid_at
        timestamptz created_at
    }
    maintenance_tickets {
        bigint id PK
        bigint tenant_id FK
        bigint room_id FK
        string title
        text description
        string priority "low | medium | high"
        string status "open | in_progress | resolved"
        int version "optimistic lock"
        timestamptz created_at
        timestamptz updated_at
        timestamptz resolved_at
    }
```

---

## Pencegahan Race Condition

| Risiko | Skenario | Mitigasi | Letak |
|---|---|---|---|
| **Double booking** kamar | Dua request menetapkan kamar `available` yang sama bersamaan | Transaction + `SELECT ... FOR UPDATE` pada baris kamar; cadangan: indeks unik parsial `leases_one_active_per_room` | `lease_service.go`, `003_leases.sql` |
| **Tagihan ganda** per periode | Owner menerbitkan tagihan periode yang sama dua kali | `UNIQUE (lease_id, billing_period)` → 409 `INVOICE_ALREADY_EXISTS` | `004_invoices.sql` |
| **Double submit** pembayaran | Penghuni klik konfirmasi bayar dua kali | `UPDATE ... WHERE status IN ('unpaid','overdue')` — hanya satu request yang mengenai baris | `invoice_repository.go` |
| **Lost update** tiket | Dua pengelola mengubah status tiket yang sama bersamaan | Optimistic locking: `UPDATE ... WHERE id=$1 AND version=$3`; versi basi → 409 `VERSION_CONFLICT` | `maintenance_repository.go` |

---

## Daftar Endpoint

Awalan `/api/v1` kecuali `/health`. Semua response berbentuk `{"success": bool, "message": "...", "data": ..., "error": {...}}`.

| Method | Path | Akses | Keterangan |
|---|---|---|---|
| GET | `/health` | Publik | Cek server dan database (di luar `/api/v1`) |
| POST | `/auth/register` | Publik | Registrasi (selalu tenant) |
| POST | `/auth/login` | Publik | Login, mengembalikan JWT (access token) |
| GET | `/auth/me` | Login | Profil pengguna |
| PATCH | `/auth/me` | Login | Ubah nama dan telepon |
| POST | `/auth/change-password` | Login | Ganti password (butuh password lama) |
| GET | `/users?role=tenant` | Owner | Daftar pengguna |
| GET | `/rooms` | Login | Daftar kamar (`?status=`) |
| GET | `/rooms/:id` | Login | Detail kamar |
| POST | `/rooms` | Owner | Tambah kamar |
| PATCH | `/rooms/:id` | Owner | Ubah kamar (parsial) |
| DELETE | `/rooms/:id` | Owner | Hapus kamar (ditolak bila ada kontrak) |
| POST | `/leases` | Owner | Buat kontrak (transaction + row lock) |
| POST | `/leases/:id/end` | Owner | Akhiri kontrak; kamar kembali `available` |
| GET | `/leases/:id` | Owner / pemilik | Detail kontrak |
| GET | `/leases` | Owner | Semua kontrak (`?status=`) |
| GET | `/leases/mine` | Tenant | Kontrak milik sendiri |
| POST | `/invoices` | Owner | Terbitkan tagihan |
| GET | `/invoices` | Owner | Semua tagihan (`?status=`) |
| GET | `/invoices/mine` | Tenant | Tagihan milik sendiri |
| GET | `/invoices/:id` | Owner / pemilik | Detail tagihan |
| POST | `/invoices/:id/pay` | Tenant | Konfirmasi pembayaran → `pending_verification` |
| POST | `/invoices/:id/verify` | Owner | Verifikasi pembayaran → `paid` |
| POST | `/invoices/:id/reject` | Owner | Tolak bukti bayar dengan alasan → `unpaid` |
| POST | `/invoices/mark-overdue` | Owner | Tandai tagihan lewat jatuh tempo → `overdue` |
| POST | `/tickets` | Tenant | Buat keluhan (kamar diambil dari kontrak aktif) |
| GET | `/tickets/mine` | Tenant | Keluhan milik sendiri |
| GET | `/tickets` | Owner | Semua keluhan (`?status=`) |
| GET | `/tickets/:id` | Owner / pemilik | Detail tiket |
| PATCH | `/tickets/:id/status` | Owner | Ubah status; wajib sertakan `version` |
| GET | `/dashboard/summary` | Owner | Ringkasan kamar, kontrak, tagihan, pendapatan bulan ini, tiket |
| GET | `/swagger/index.html` | Publik* | Swagger UI (*hanya bila `SWAGGER_ENABLED=true`, selain itu 404; di luar `/api/v1`) |

---

## Stack Teknologi

| Kebutuhan | Teknologi |
|---|---|
| Bahasa | Go 1.24 |
| HTTP | Fiber v2 |
| Database | PostgreSQL via pgx v5 (`pgxpool`) |
| Autentikasi | JWT HS256 + bcrypt |
| Konfigurasi | godotenv |
| Container | Docker + Docker Compose |
| Dokumentasi API | swaggo (`swag` v1.16.6) + `gofiber/swagger` v1.1.1 |

---

## Cara Menjalankan

Prasyarat: **Go** dan **PostgreSQL** (lokal atau Supabase).

```powershell
# 1. Salin file konfigurasi
Copy-Item .env.example .env

# 2. Isi .env — wajib diisi:
#    DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME (atau DATABASE_URL)
#    JWT_SECRET (minimal 32 karakter)
#    SEED_OWNER_EMAIL, SEED_OWNER_PASSWORD (untuk akun owner pertama)
notepad .env

# 3. Buat database
psql -U postgres -c "CREATE DATABASE hunispace;"

# 4. Jalankan (migration otomatis saat AUTO_MIGRATE=true)
go mod tidy
go test ./...
go run .
```

Migration dijalankan otomatis saat server menyala (`AUTO_MIGRATE=true`). Untuk membuat akun owner pertama, isi `SEED_OWNER_EMAIL` dan `SEED_OWNER_PASSWORD` di `.env`.

### Dengan Docker (database ikut dijalankan)

Prasyarat: Docker Desktop. Di `.env` isi `DB_PASSWORD` (wajib), `JWT_SECRET` (minimal 32 karakter), `SEED_OWNER_EMAIL`, dan `SEED_OWNER_PASSWORD`. Opsional: `SWAGGER_ENABLED=true` dan `HOST_PORT` (bawaan 8080).

```powershell
Copy-Item .env.example .env
docker compose up --build
```

Aplikasi berjalan di `http://localhost:8080` (port 8080 di host diteruskan ke port 3000 di container; cek `http://localhost:8080/health`; Swagger di `/swagger/index.html` bila diaktifkan) dan PostgreSQL berjalan di container `db` (data tersimpan di volume `hunispace-data`). Berhenti: `Ctrl+C`, lalu `docker compose down` (tambahkan `-v` untuk menghapus data).

### Swagger UI

Dokumentasi dibuat dengan swaggo (code-first): anotasi `@Summary`, `@Router`, dan seterusnya ditulis di atas method di `app/service/`, informasi umum di `main.go`. Hasil generate ada di folder `docs/` (`docs.go`, `swagger.json`, `swagger.yaml`).

Pasang sekali (butuh internet):

```powershell
go get -tool github.com/swaggo/swag/cmd/swag@v1.16.6
go get github.com/gofiber/swagger@v1.1.1
go mod tidy
```

Hasilkan dokumentasi, ulangi setiap kali anotasi berubah:

```powershell
go tool swag init
```

Swagger UI hanya menyala bila `SWAGGER_ENABLED=true` di `.env` (bawaan `false`, hasilnya 404). Setelah server berjalan, buka `http://localhost:3000/swagger/index.html`, klik **Authorize**, isi `Bearer <access_token>` (kata `Bearer`, satu spasi, lalu token dari `POST /auth/login`), lalu coba endpoint dari halaman. Penjelasan lengkap ada di [docs/api-documentation.md](docs/api-documentation.md).

### Uji Cepat

Cek server berjalan:

```powershell
Invoke-RestMethod http://localhost:3000/health
```

Uji registrasi dan login:

```powershell
$body = @{ full_name = "Budi"; email = "budi@example.com"; password = "rahasia123" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri http://localhost:3000/api/v1/auth/register -ContentType "application/json" -Body $body

$login = @{ email = "budi@example.com"; password = "rahasia123" } | ConvertTo-Json
$res = Invoke-RestMethod -Method Post -Uri http://localhost:3000/api/v1/auth/login -ContentType "application/json" -Body $login

# Akses endpoint dengan token
Invoke-RestMethod -Uri http://localhost:3000/api/v1/rooms -Headers @{ Authorization = "Bearer $($res.data.access_token)" }
```

---

## Dokumentasi Tim

| Dokumen | Isi |
|---|---|
| [docs/development-workflow.md](docs/development-workflow.md) | Setup, alur request, aturan Git dan branch, konvensi commit |
| [docs/sop-create-module.md](docs/sop-create-module.md) | Langkah baku membuat modul baru |
| [docs/database-schema.md](docs/database-schema.md) | Skema database dan constraint |
| [docs/api-documentation.md](docs/api-documentation.md) | Dokumentasi API dan penggunaan Swagger |

---

## Alur Git

```
main (commit awal) → feat/core-auth (Zelvia, PR #1)
                   → feat/room-lease (Drefita, PR #2)
                   → feat/invoice-maintenance (Putri, PR #3)
                   → docs/project-docs (Zelvia, PR #4)
```

- Setiap anggota bekerja di branch sendiri dengan **micro commit** (1 commit = 1 perubahan kecil).
- Semua perubahan ke `main` wajib lewat **Pull Request**.
- Git Master mereview (`go build`, `go vet`, `go test`) lalu merge dengan **Create a merge commit**, bukan Squash. Hal tersebut agar commit tiap anggota tetap terlihat di riwayat.
- Urutan merge mengikuti ketergantungan: fondasi → rooms+leases → invoices+maintenance.

Konvensi commit: `type(scope): deskripsi`. Contoh:
```
feat(lease): add lease service with transaction and row lock against double booking
test(invoice): add table-driven tests for billing period validation
docs(readme): add project description, features, roles, and database design
```