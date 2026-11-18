# Development Workflow

Panduan cara kerja harian tim HuniSpace. Baca sebelum mengerjakan modul pertamamu.

## 1. Stack

| Kebutuhan | Teknologi |
|---|---|
| Bahasa | Go |
| HTTP framework | Fiber v2 |
| Database | PostgreSQL via pgx v5 (`pgxpool`) |
| Autentikasi | JWT HS256 (access token) + bcrypt |
| Config | godotenv |

## 2. Setup lokal

```powershell
git clone <url-repo>
cd hunispace
Copy-Item .env.example .env       # isi DB_PASSWORD (atau DATABASE_URL) dan JWT_SECRET
psql -U postgres -c "CREATE DATABASE hunispace;"
go mod download
go run .
```

Jalankan dari **root project**. `.env` tidak boleh di-commit.

## 3. Arsitektur dan alur request

```text
route → service → repository → database
```

| Lapisan | Tugas | TIDAK boleh |
|---|---|---|
| `route` | Mendaftarkan path dan memasang middleware | Berisi logika bisnis |
| `service` | Membaca input, validasi, aturan bisnis, membungkus response | Menulis SQL langsung |
| `repository` | Query database | Berisi aturan bisnis atau mengenal Fiber |
| `model` | Struct dan konstanta | Berisi logika selain hal sederhana |

Tidak ada lapisan pengendali HTTP yang terpisah. Fungsi HTTP berada di `service`.

Contoh alur `POST /api/v1/leases`:
1. `route/lease_route.go` mengarahkan ke `LeaseService.Create` setelah middleware `Auth` dan `Owner`.
2. `Auth` memvalidasi JWT dan menyimpan identitas di context.
3. `Create` membaca body, memvalidasi (`lease_rules.go`), lalu membuka transaction.
4. Di dalam transaction: kunci baris kamar, periksa status, buat kontrak, ubah status kamar.
5. `repository` hanya menjalankan SQL; `service` membungkus hasil dengan `helper.Created`.

## 4. Kontrak response

Semua endpoint memakai bentuk yang sama; gunakan helper, jangan menyusun JSON sendiri.

```go
return helper.Success(c, fiber.StatusOK, "pesan sukses", data)
return helper.Created(c, "pesan sukses", data)
return helper.Conflict("KODE_SPESIFIK", "pesan untuk klien")
```

| Aksi | Sukses | Gagal umum |
|---|---|---|
| Create | 201 | 422 validasi, 409 konflik |
| Get / List | 200 | 404 tidak ada |
| Update | 200 | 422, 404, 409 |
| Delete | 200 | 404, 409 bila masih dipakai |
| Tanpa token / token salah | | 401 |
| Role tidak berhak | | 403 |

## 5. Konvensi kode

- Satu berkas per modul di tiap lapisan: `room.go`, `room_repository.go`, `room_service.go`, `room_route.go`.
- Aturan yang dapat diuji tanpa database ditulis sebagai fungsi murni di `*_rules.go` dan diuji dengan test table-driven.
- Repository berbentuk interface + struct kecil (`NewRoomRepository()`), setiap method menerima `DBTX` agar dapat dipakai di dalam transaction.
- `user_id` selalu diambil dari token (`helper.CurrentUser`), bukan dari body.
- Query data milik pengguna wajib menyertakan filter pemilik.
- Jalankan `gofmt -w .`, `go vet ./...`, dan `go test ./...` sebelum commit.

## 6. Git dan branching

| Branch | Peran | Siapa |
|---|---|---|
| `main` | Hasil akhir | **Hanya Git Master** yang merge |
| `feat/<nama-modul>` | Branch kerja anggota | Pemilik branch |

Branch kerja bersifat **sementara**: satu anggota memegang satu branch aktif pada satu waktu. Setelah Pull Request di-merge, hapus branch-nya dan buat yang baru dari `main`. Riwayat commit **tidak hilang** karena sudah tercatat di `main`.

```powershell
git checkout main
git pull origin main
git checkout -b feat/room-lease
# ... kerja, commit kecil ...
git push -u origin feat/room-lease
```

Lalu buka Pull Request di GitHub (base `main`, compare branch-mu) dan kabari Git Master.

### Aturan agar tidak bentrok
1. `main.go`, `go.mod`, `go.sum`, dan `.env.example` hanya diubah Git Master. Butuh dependensi baru? Minta ke Git Master.
2. Setiap modul mendaftarkan route-nya sendiri lewat `init()`; jangan mengedit `route.go`.
3. Nomor migration sudah dipesan per modul (001–005). Jangan membuat nomor lain tanpa kesepakatan.
4. Sebelum mulai bekerja: `git pull origin main`, lalu `git merge main` ke branch-mu bila `main` berubah.

### Konvensi commit

Satu commit satu hal. Format `type(scope): deskripsi`.

```text
feat(lease): add lease service with tx and row lock against double booking
test(invoice): add table-driven tests for billing period validation
fix(room): reject status change while room is occupied
docs: add database schema
```

Tipe: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`.

### Merge oleh Git Master

Di halaman Pull Request, pilih **"Create a merge commit"**. **Jangan** pilih "Squash and merge": itu menggabungkan semua micro commit menjadi satu dan menghilangkan keaktifan commit tiap anggota.

Sebelum merge, Git Master memeriksa:

```powershell
git fetch origin
git checkout feat/room-lease
git pull
go build ./...
go vet ./...
go test ./...
```

## 7. Checklist sebelum Pull Request

- [ ] Mengikuti lapisan route → service → repository → model.
- [ ] Response memakai `helper.Success` / `helper.Created` / error dari `helper`.
- [ ] Endpoint di balik `d.Auth`, dan `d.Owner` / `d.Tenant` bila perlu.
- [ ] Tidak ada `.env` atau kredensial yang ter-commit.
- [ ] `gofmt`, `go build`, `go vet`, `go test` lulus.
- [ ] Pesan commit mengikuti format.

## 8. Keamanan

- `.env` ada di `.gitignore`; bagikan lewat kanal aman, bukan repository.
- Login menjawab sama persis untuk email tidak ada dan password salah.
- Error tak terduga hanya dicatat di log; klien menerima pesan 500 generik.
