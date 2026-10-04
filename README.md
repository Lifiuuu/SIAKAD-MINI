# SIAKAD Mini — Backend API

Backend REST API sederhana untuk Sistem Informasi Akademik (SIAKAD), dibangun dengan Go + Fiber.

## Tech Stack

- **Framework**: [Fiber v2](https://gofiber.io/)
- **Database**: PostgreSQL (via [pgx/v5](https://github.com/jackc/pgx))
- **Auth**: JWT (Access Token + Refresh Token)
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator)
- **Hashing**: bcrypt
- **Logger**: `log/slog` + lumberjack

## Struktur Direktori

```
SIAKAD_MINI/
├── app/
│   ├── model/          # Struct entitas & request/response
│   ├── repository/     # Interface + implementasi query SQL
│   └── service/        # Business logic + handler (menerima *fiber.Ctx)
├── config/             # App config, env, logger
├── database/           # Connection pool PostgreSQL
├── helper/             # JWT, bcrypt, validator, cursor, response, dll.
├── middleware/         # Auth, authz, CORS, rate limiter
├── migrations/         # SQL migration files
├── route/              # Pendaftaran route
├── main.go             # Entry point
├── .env.example
└── go.mod
```

## Cara Menjalankan

### 1. Persiapkan database

```bash
createdb siakad_mini

# Jalankan migration secara berurutan
psql -U postgres -d siakad_mini -f migrations/001_create_users.sql
psql -U postgres -d siakad_mini -f migrations/002_create_mahasiswa.sql
psql -U postgres -d siakad_mini -f migrations/003_create_mata_kuliah.sql
psql -U postgres -d siakad_mini -f migrations/004_create_nilai.sql
psql -U postgres -d siakad_mini -f migrations/005_auth.sql
psql -U postgres -d siakad_mini -f migrations/006_rbac.sql
```

### 2. Buat file `.env`

Salin `.env.example` menjadi `.env` dan isi sesuai konfigurasi lokal:

```bash
cp .env.example .env
```

Isi minimal yang wajib:
```
DB_PASSWORD=password_postgres_anda
JWT_SECRET=minimal_32_karakter_acak_dan_panjang
```

### 3. Jalankan

```bash
go run main.go
```

## Endpoint API

### Autentikasi (`/api/v1/auth`)

| Method | Path | Keterangan |
|--------|------|------------|
| POST | `/auth/register` | Daftar akun baru |
| POST | `/auth/login` | Login, mendapatkan token |

| GET | `/auth/me` | Profil pengguna saat ini |

### User (`/api/v1/users`) — Butuh login

| Method | Path | Permission |
|--------|------|------------|
| GET | `/users/` | `user:list` |
| POST | `/users/` | `user:update:any` |
| GET | `/users/:id` | Own / `user:read:any` |
| PUT | `/users/:id` | Own / `user:update:any` |
| PATCH | `/users/:id` | Own / `user:update:any` |
| DELETE | `/users/:id` | `user:delete` |
| PATCH | `/users/:id/role` | `role:assign` |

### Mahasiswa (`/api/v1/mahasiswa`) — Butuh login

| Method | Path | Permission |
|--------|------|------------|
| GET | `/mahasiswa/` | `mahasiswa:list` |
| POST | `/mahasiswa/` | `mahasiswa:create` |
| GET | `/mahasiswa/:id` | Own / `mahasiswa:read:any` |
| PUT | `/mahasiswa/:id` | Own / `mahasiswa:update:any` |
| PATCH | `/mahasiswa/:id` | Own / `mahasiswa:update:any` |
| DELETE | `/mahasiswa/:id` | `mahasiswa:delete` |

### Mata Kuliah (`/api/v1/mata-kuliah`) — Butuh login

| Method | Path | Permission |
|--------|------|------------|
| GET | `/mata-kuliah/` | `matakuliah:list` |
| POST | `/mata-kuliah/` | `matakuliah:create` |
| GET | `/mata-kuliah/:id` | `matakuliah:list` |
| PUT | `/mata-kuliah/:id` | `matakuliah:update:any` |
| PATCH | `/mata-kuliah/:id` | `matakuliah:update:any` |
| DELETE | `/mata-kuliah/:id` | `matakuliah:delete` |

### Nilai (`/api/v1/nilai`) — Butuh login

| Method | Path | Keterangan |
|--------|------|------------|
| GET | `/nilai/?nim=NIM` | Ambil semua nilai berdasarkan NIM |
| POST | `/nilai/` | Input nilai baru |
| GET | `/nilai/:id` | Detail nilai |
| PUT | `/nilai/:id` | Ganti nilai (perlu `nilai:update:any`) |
| PATCH | `/nilai/:id` | Ubah sebagian nilai (perlu `nilai:update:any`) |
| DELETE | `/nilai/:id` | Hapus nilai (perlu `nilai:delete`) |

### Health Check

| Method | Path | Keterangan |
|--------|------|------------|
| GET | `/api/v1/health` | Cek status server & database |

## RBAC — Role-Based Access Control

| Role | Hak Akses |
|------|-----------|
| `admin` | Akses penuh ke semua resource |
| `dosen` | Kelola mahasiswa & nilai |
| `mahasiswa` | Hanya lihat data (mahasiswa, mata kuliah, nilai sendiri) |
| `user` | Default saat registrasi, tidak ada permission khusus |

## Format Respons

Semua response menggunakan amplop standar:

```json
{
  "success": true,
  "message": "pesan deskriptif",
  "data": { ... },
  "meta": { "limit": 10, "has_more": false }
}
```

Error response:

```json
{
  "success": false,
  "code": "VALIDATION_ERROR",
  "message": "validasi gagal",
  "fields": { "nim": "format NIM tidak valid" },
  "request_id": "abc123"
}
```

## Konversi Grade Nilai

| Nilai | Grade |
|-------|-------|
| ≥ 85  | A     |
| ≥ 75  | B+    |
| ≥ 70  | B     |
| ≥ 60  | C+    |
| ≥ 55  | C     |
| ≥ 40  | D     |
| < 40  | E     |
