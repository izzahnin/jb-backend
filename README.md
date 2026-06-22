# jb-backend

Go REST API untuk sistem manajemen fleet dan order PT. Jalur Berlian Makassar.

---

## Quick Start

```bash
# Jalankan semua service (PostgreSQL + Redis + API)
docker-compose up -d --build

# API: http://localhost:8080
# Swagger: http://localhost:8080/swagger/index.html
```

> **Pertama kali setup:** database dibuat otomatis dari `migrations/init.sql` saat container PostgreSQL pertama kali start.
>
> **Reset database:** `docker-compose down -v && docker-compose up -d --build`

---

## Demo / Seed Data

Untuk demo dan portfolio, jalankan seed data setelah database jalan:

```bash
docker exec -i jbm_postgres psql -U admin -d jalur_berlian_db < migrations/seed.sql
```

**Yang di-insert:**

| Data | Jumlah | Detail |
|------|--------|--------|
| Users | 3 | superadmin · admin.sales · admin.ops |
| Customers | 3 | PT. Nusantara Logistik, CV. Maju Bersama, PT. Sulawesi Cargo |
| Trucks | 3 | Fuso Box, Tronton, Trailer — status bervariasi |
| Drivers | 3 | Status bervariasi (available / on_duty) |
| Orders | 5 | 2 completed, 1 partial (in_transit), 2 pending |
| Trips | 3 | 2 delivered, 1 in_transit dengan rute GPS Makassar→Parepare |
| Locations | 9 | Titik GPS realistis sepanjang Trans-Sulawesi |

**Login setelah seed:**

| Username | Password | Role |
|----------|----------|------|
| `superadmin` | `demo1234` | super_admin |
| `admin.sales` | `demo1234` | admin_sales |
| `admin.ops` | `demo1234` | admin_ops |

> ⚠️ Seed menghapus semua data sebelumnya. Jalankan hanya di environment demo.

---

## Environment

Buat file `.env.local` (copy dari `.env.local.example`):

```
DB_SOURCE=postgres://admin:password123@localhost:5432/jalur_berlian_db?sslmode=disable
REDIS_ADDR=localhost:6379
JWT_SECRET=dev-secret-key-change-in-production
PORT=8080
GIN_MODE=debug
```

---

## Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| Language | Go 1.25 |
| Framework | Gin v1.12 |
| Database | PostgreSQL 15 |
| Cache | Redis 7 |
| Auth | JWT HS256 (24h) |
| ORM | sqlx v1.4 |

---

## Architecture

```
Handler → Usecase → Repository → PostgreSQL / Redis
```

Clean architecture. Semua dependency di-inject di `cmd/api/main.go`. Tidak ada global state.

---

## Project Structure

```
jb-backend/
├── cmd/api/main.go              Entry point + dependency injection
├── internal/
│   ├── handler/                 HTTP handlers (12 file, per domain)
│   │   ├── handler.go           Struct Handler + semua dependency
│   │   ├── routes.go            Registrasi semua route
│   │   ├── auth_handler.go      POST /auth/login, /auth/logout
│   │   ├── public_routes.go     GET /track/:order_number (tanpa auth)
│   │   ├── user_routes.go       CRUD /admin/users
│   │   ├── customer_routes.go   CRUD /admin/customers
│   │   ├── driver_routes.go     CRUD /admin/drivers
│   │   ├── truck_routes.go      CRUD /admin/trucks
│   │   ├── order_routes.go      CRUD /admin/orders
│   │   ├── trip_routes.go       CRUD + status /admin/trips
│   │   ├── location_routes.go   GPS /admin/trips/:id/location
│   │   └── dashboard_routes.go  GET /admin/dashboard/stats
│   ├── usecase/                 Business logic + validasi
│   ├── repository/              Query database (sqlx)
│   ├── model/                   Struct entity + DTO request
│   └── middleware/auth.go       JWT + RBAC middleware
├── pkg/
│   ├── database/postgres.go     Init PostgreSQL pool
│   ├── database/redis.go        Init Redis client
│   └── helper/pagination.go     Helper limit/offset
└── migrations/
    ├── init.sql                 Schema lengkap (dijalankan Docker saat init)
    └── 000010_add_timestamps.*  Migration tambahan (referensi)
```

---

## Endpoints

### Auth (publik)
| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | `/auth/login` | Login, dapat JWT token |
| POST | `/auth/logout` | Logout |

### Admin (butuh JWT)
| Method | Path | Akses |
|--------|------|-------|
| GET/POST | `/admin/customers` | super_admin, admin_sales |
| GET/PATCH/DELETE | `/admin/customers/:id` | super_admin, admin_sales |
| GET/POST | `/admin/drivers` | super_admin, admin_ops |
| GET/PATCH/DELETE | `/admin/drivers/:id` | super_admin, admin_ops |
| GET/POST | `/admin/trucks` | super_admin, admin_ops |
| GET/PATCH/DELETE | `/admin/trucks/:id` | super_admin, admin_ops |
| GET/POST | `/admin/orders` | super_admin, admin_sales |
| GET/PATCH/DELETE | `/admin/orders/:id` | super_admin, admin_sales |
| GET/POST | `/admin/trips` | super_admin, admin_ops |
| PATCH | `/admin/trips/:id/status` | super_admin, admin_ops |
| POST | `/admin/trips/:id/location` | super_admin, admin_ops |
| GET | `/admin/trips/:id/locations` | super_admin, admin_ops |
| GET/POST/DELETE | `/admin/users` | super_admin |
| GET | `/admin/dashboard/stats` | semua admin |

### Publik (tanpa auth)
| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/track/:order_number` | Tracking order publik |

---

## Domain & State Machine

**Order status:** `pending → partial → completed | cancelled`

**Trip status:** `pickup → in_transit → delivered | cancelled`

Status hanya maju, tidak bisa mundur.

---

## Key Patterns

- **Soft delete** — semua entity punya `is_active bool`. DELETE = set `is_active = false`
- **Audit trail** — field `created_by`, `updated_by` (FK ke users) + `created_by_name`, `updated_by_name` di-JOIN dari users
- **Timestamps** — semua tabel punya `created_at` + `updated_at`. `updated_at` di-set `CURRENT_TIMESTAMP` saat UPDATE
- **GPS cache** — posisi terbaru trip di Redis, history di PostgreSQL (tabel `locations`)
- **Pagination** — semua list endpoint support `?limit=&offset=`

---

## Database Schema (Kolom Penting)

| Tabel | Kolom audit |
|-------|-------------|
| customers | `created_at`, `updated_at`, `created_by`, `updated_by` |
| drivers | `created_at`, `updated_at`, `created_by`, `updated_by` |
| trucks | `created_at`, `updated_at`, `created_by`, `updated_by` |
| orders | `created_at`, `updated_at`, `order_date`, `admin_id` |
| trips | `created_at`, `start_time`, `end_time`, `created_by`, `started_by`, `completed_by` |

---

## Testing

Tidak ada test suite otomatis. Test manual via:
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Postman**: import `Jalur-Berlian-Backend.postman_collection.json`

---

**Author:** Nurul Izzah Nurhidayat | **Last Updated:** 2026-06
