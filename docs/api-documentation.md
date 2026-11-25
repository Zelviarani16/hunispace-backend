# Dokumentasi API HuniSpace

Basis URL: `http://localhost:3000`. Semua endpoint di bawah `/api/v1` kecuali `/health`. Kontrak interaktif (skema, kode error) tersedia di Swagger UI (`/swagger/index.html`, hanya bila `SWAGGER_ENABLED=true`) dan di `docs/swagger.json` hasil `go tool swag init`; berkas JSON itu dapat diimpor ke Postman atau editor.swagger.io.

## Konvensi

- Autentikasi: header `Authorization: Bearer <access_token>` dari `POST /auth/login`.
- Sukses: `{"success": true, "message": "...", "data": ...}`.
- Gagal: `{"success": false, "message": "...", "error": {"code": "...", "fields": {...}}}`.
- Kode status: 200/201 sukses, 400 body salah, 401 belum login, 403 role tidak cukup, 404 tidak ditemukan (juga untuk data milik orang lain), 409 konflik, 422 validasi gagal.
- Tanggal `YYYY-MM-DD`, periode tagihan `YYYY-MM`, uang dalam Rupiah (angka).

## Daftar endpoint


### System

| Method | Path | Akses | Fungsi | Body | Kemungkinan error |
|---|---|---|---|---|---|
| GET | `/health` | Publik | Cek server dan database | - | - |

### Auth

| Method | Path | Akses | Fungsi | Body | Kemungkinan error |
|---|---|---|---|---|---|
| POST | `/auth/register` | Publik | Registrasi akun (selalu menjadi tenant) | `RegisterRequest` | 409 |
| POST | `/auth/login` | Publik | Login, mengembalikan access token | `LoginRequest` | 401 |
| GET | `/auth/me` | Login | Profil pengguna yang login | - | - |
| PATCH | `/auth/me` | Login | Ubah nama dan telepon | `UpdateProfileRequest` | - |
| POST | `/auth/change-password` | Login | Ganti password | `ChangePasswordRequest` | 401 |

### Users

| Method | Path | Akses | Fungsi | Body | Kemungkinan error |
|---|---|---|---|---|---|
| GET | `/users` | Owner | Daftar pengguna | - | - |

### Rooms

| Method | Path | Akses | Fungsi | Body | Kemungkinan error |
|---|---|---|---|---|---|
| GET | `/rooms` | Login | Daftar kamar | - | - |
| POST | `/rooms` | Owner | Tambah kamar | `RoomCreateRequest` | 409 |
| GET | `/rooms/{id}` | Login | Detail kamar | - | 404 |
| PATCH | `/rooms/{id}` | Owner | Ubah kamar (parsial) | `RoomUpdateRequest` | 404, 422 |
| DELETE | `/rooms/{id}` | Owner | Hapus kamar | - | 404, 409 |

### Leases

| Method | Path | Akses | Fungsi | Body | Kemungkinan error |
|---|---|---|---|---|---|
| POST | `/leases` | Owner | Buat kontrak sewa | `LeaseCreateRequest` | 400, 404, 409 |
| GET | `/leases` | Owner | Semua kontrak | - | - |
| GET | `/leases/mine` | Tenant | Kontrak milik sendiri | - | - |
| GET | `/leases/{id}` | Login | Detail kontrak | - | 404 |
| POST | `/leases/{id}/end` | Owner | Akhiri kontrak, kamar kembali available | - | 404, 409 |

### Invoices

| Method | Path | Akses | Fungsi | Body | Kemungkinan error |
|---|---|---|---|---|---|
| POST | `/invoices` | Owner | Terbitkan tagihan | `InvoiceCreateRequest` | 409 |
| GET | `/invoices` | Owner | Semua tagihan | - | - |
| GET | `/invoices/mine` | Tenant | Tagihan milik sendiri | - | - |
| POST | `/invoices/mark-overdue` | Owner | Tandai tagihan lewat jatuh tempo | - | - |
| GET | `/invoices/{id}` | Login | Detail tagihan | - | 404 |
| POST | `/invoices/{id}/pay` | Tenant | Konfirmasi pembayaran | `InvoicePayRequest` | 409 |
| POST | `/invoices/{id}/verify` | Owner | Verifikasi pembayaran (menjadi paid) | - | 409 |
| POST | `/invoices/{id}/reject` | Owner | Tolak bukti bayar dengan alasan (kembali unpaid) | `InvoiceRejectRequest` | 409 |

### Tickets

| Method | Path | Akses | Fungsi | Body | Kemungkinan error |
|---|---|---|---|---|---|
| POST | `/tickets` | Tenant | Buat tiket keluhan | `TicketCreateRequest` | 409 |
| GET | `/tickets` | Owner | Semua tiket | - | - |
| GET | `/tickets/mine` | Tenant | Tiket milik sendiri | - | - |
| GET | `/tickets/{id}` | Login | Detail tiket | - | 404 |
| PATCH | `/tickets/{id}/status` | Owner | Ubah status tiket | `TicketStatusRequest` | 409 |

### Dashboard

| Method | Path | Akses | Fungsi | Body | Kemungkinan error |
|---|---|---|---|---|---|
| GET | `/dashboard/summary` | Owner | Ringkasan untuk owner | - | - |

## Contoh request

**LoginRequest**

```json
{
  "email": "budi@example.com",
  "password": "rahasia123"
}
```

**RoomCreateRequest**

```json
{
  "room_number": "A-101",
  "room_type": "Single",
  "monthly_price": 1500000,
  "facilities": "AC, kasur"
}
```

**LeaseCreateRequest**

```json
{
  "tenant_id": 2,
  "room_id": 1,
  "start_date": "2026-11-01",
  "end_date": "2027-05-01"
}
```

**InvoiceCreateRequest**

```json
{
  "lease_id": 1,
  "billing_period": "2026-11",
  "amount": 1500000,
  "due_date": "2026-12-01"
}
```

**InvoicePayRequest**

```json
{
  "payment_method": "transfer",
  "proof_url": "https://contoh.com/bukti.jpg"
}
```

**InvoiceRejectRequest**

```json
{
  "note": "Bukti transfer buram"
}
```

**TicketCreateRequest**

```json
{
  "title": "AC bocor",
  "description": "AC meneteskan air ke kasur",
  "priority": "medium"
}
```

**TicketStatusRequest**

```json
{
  "status": "in_progress",
  "version": 1
}
```

## Aturan bisnis yang terlihat di API

| Situasi | Hasil |
|---|---|
| Dua penyewa memesan kamar yang sama | 409 untuk yang kedua (transaksi + row lock + unique index parsial) |
| Tagihan periode sama dibuat dua kali | 409 (`UNIQUE (lease_id, billing_period)`) |
| Tombol bayar ditekan dua kali | 409 untuk klik kedua (UPDATE bersyarat) |
| Update tiket dengan `version` usang | 409 (optimistic locking) |
| Tenant membuka data orang lain | 404 |
| Tenant memanggil endpoint owner | 403 |
