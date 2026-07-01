# CLAUDE.md — jb-backend

Go REST API untuk sistem manajemen fleet PT. Jalur Berlian Makassar.

---

## Running

```bash
# Docker (recommended — nyalain PostgreSQL + Redis + API sekaligus)
docker-compose up -d --build      # wajib --build setelah ada perubahan kode Go

# Reset database bersih (hapus semua data)
docker-compose down -v && docker-compose up -d --build

# Development tanpa Docker (butuh PostgreSQL & Redis jalan duluan)
go run cmd/api/main.go

# Stop Docker
docker-compose down
```

API jalan di **port 8080**. Swagger UI: `http://localhost:8080/swagger/docs`

## Environment

Copy `.env.local.example` → `.env.local` lalu isi secrets:

```
DB_SOURCE              postgres://user:pass@localhost:5432/jalur_berlian_db?sslmode=disable
REDIS_ADDR             localhost:6379
REDIS_PASSWORD                            # kosong untuk lokal, isi untuk Upstash/production
JWT_SECRET             [32-byte hex: openssl rand -hex 32]
PORT                   8080
GIN_MODE               debug              # set ke "release" di production
CORS_ALLOWED_ORIGINS   http://localhost:3000  # pisah koma untuk multiple origin
```

App load `.env.local` dulu (secrets), baru `.env` (defaults).

**Production env vars tambahan:**
- `CORS_ALLOWED_ORIGINS` → `https://jalurberlian.vercel.app,https://jalurberlian.id`
- `GIN_MODE` → `release`
- `DB_SOURCE` → pakai `sslmode=require` di Render/Railway

## Build & Lint

```bash
go build ./...
go vet ./...
```

## Testing

Tidak ada test suite. Test manual via:
- Swagger UI di `http://localhost:8080/swagger/index.html`
- Postman: `Jalur-Berlian-Backend.postman_collection.json`

---

## Architecture

```
cmd/api/main.go   →   Handler   →   Usecase   →   Repository   →   PostgreSQL
                                                              ↘   Redis (GPS cache)
```

Clean architecture. Semua dependency di-inject di `main.go` startup. Urutan init:
1. Load env → config
2. Connect PostgreSQL (pool 25 conn) + Redis
3. Init 8 repository
4. Init 8 usecase (inject repo ke usecase)
5. Init handler (inject semua usecase)
6. Register routes
7. Start server port 8080 dengan graceful shutdown (SIGINT/SIGTERM, timeout 10s)

---

## File Map

### Handler — `internal/handler/`
| File | Isi |
|------|-----|
| `handler.go` | Struct `Handler` dengan semua dependency |
| `routes.go` | Orchestrator registrasi semua route |
| `auth_handler.go` | POST /login, POST /logout |
| `public_routes.go` | GET /track/:order_number (tanpa auth) |
| `user_routes.go` | CRUD /users (super_admin only) |
| `customer_routes.go` | CRUD /customers |
| `driver_routes.go` | CRUD /drivers |
| `truck_routes.go` | CRUD /trucks |
| `order_routes.go` | CRUD /orders |
| `trip_routes.go` | CRUD + status /trips |
| `location_routes.go` | POST/GET /locations (GPS) |
| `dashboard_routes.go` | GET /dashboard/stats |

### Usecase — `internal/usecase/`
`auth_usecase.go`, `user_usecase.go`, `customer_usecase.go`, `driver_usecase.go`, `truck_usecase.go`, `order_usecase.go`, `trip_usecase.go`, `location_usecase.go`, `errors.go`

### Repository — `internal/repository/`
`user_repository.go`, `customer_repository.go`, `driver_repository.go`, `truck_repository.go`, `order_repository.go`, `trip_repository.go`, `location_repository.go`, `audit_log_repository.go`

Semua repo: `GetByID`, `FetchAll`, `FetchAllWithPagination`, `Create`, `Update`, `Delete` via `sqlx`.

### Model — `internal/model/`
`auth.go`, `customer.go`, `driver.go`, `truck.go`, `order.go`, `trip.go`, `location.go`

Struct field: `db:""` (sqlx), `json:""`, Swagger `example:""` tags.

### Middleware — `internal/middleware/auth.go`
- `AuthMiddleware(jwtSecret)` — validasi JWT dari Authorization header
- `AdminMiddleware()` — cek role valid (super_admin / admin_sales / admin_ops)
- `RequireRoles(...string)` — RBAC granular per endpoint

### Pkg
- `pkg/database/postgres.go` — init PostgreSQL pool
- `pkg/database/redis.go` — init Redis client
- `pkg/helper/pagination.go` — helper limit/offset pagination

---

## Key Patterns

**Dependency injection:** semua dep di-pass ke constructor di `main.go`, tidak ada global state.

**Custom errors:** `internal/usecase/errors.go` — usecase return custom error type, handler map ke HTTP status code.

**JWT:** HS256, expiry 24 jam. Secret dari env `JWT_SECRET`.

**Soft delete:** semua entity punya field `is_active bool`. Delete = set `is_active = false`.

**Audit trail:** field `created_by_name`, `updated_by_name` di semua entity utama.

**GPS tracking:** device push koordinat secara berkala (interval-based, bukan real-time streaming). Posisi terbaru di-cache Redis untuk akses cepat; history lengkap disimpan di PostgreSQL (`location` table).

**Pagination:** semua list endpoint support `?limit=&offset=` via `pkg/helper/pagination.go`.

**CORS:** dibaca dari env var `CORS_ALLOWED_ORIGINS` (comma-separated). Default lokal: `localhost:3000,localhost:3001`. Di production set ke domain frontend yang aktif.

---

## Domain & State Machines

**Order status:** `pending → partial → completed | cancelled`

**Trip status:** `pickup → in_transit → delivered | cancelled`

Status hanya bisa maju, tidak bisa mundur.

**Role-based access:**
| Role | Akses |
|------|-------|
| `super_admin` | Semua endpoint |
| `admin_sales` | Customers, Orders |
| `admin_ops` | Trucks, Drivers, Trips, Locations |

---

## Domain Entities

- **Customer** — perusahaan pengirim (company_name, pic_name, phone, npwp)
- **Driver** — supir truk (license_number, status: available/on_duty/off)
- **Truck** — kendaraan (plate_number, truck_type, status: available/on_duty/maintenance)
- **Order** — pesanan pengiriman (1 order = 1 container, bisa banyak trip)
- **Trip** — 1 perjalanan truk per container (container_number, seal_number)
- **Location** — koordinat GPS trip (latitude, longitude, timestamp)

---

## Dependencies Utama

```
github.com/gin-gonic/gin v1.12.0
github.com/jmoiron/sqlx v1.4.0
github.com/golang-jwt/jwt/v5 v5.3.1
github.com/redis/go-redis/v9 v9.18.0
github.com/lib/pq v1.11.2
golang.org/x/crypto v0.49.0   (bcrypt)
github.com/swaggo/swag v1.16.6
```

Go version: 1.25.0

---

## Migrations

Schema utama ada di `migrations/init.sql` — dijalankan otomatis saat Docker container PostgreSQL pertama kali start (volume kosong).

**Reset database bersih:** `docker-compose down -v && docker-compose up -d --build`

File `000010_add_timestamps.*` ada sebagai referensi tapi tidak dijalankan otomatis — semua perubahannya sudah masuk ke `init.sql`.

**Seed data demo:** `migrations/seed.sql` — insert 3 user, 3 customer, 3 truck, 3 driver, 5 order, 3 trip, 9 GPS point. Jalankan dengan:
```bash
# Lokal (Docker)
docker exec -i jbm_postgres psql -U admin -d jalur_berlian_db < migrations/seed.sql

# Production (Render/Railway)
psql "<DATABASE_URL>" < migrations/seed.sql
```
Login seed: `superadmin` / `admin.sales` / `admin.ops` — password: `demo1234`

### Kolom audit di setiap tabel

| Tabel | created_at | updated_at | created_by | updated_by | Catatan lain |
|-------|-----------|-----------|-----------|-----------|--------------|
| customers | ✅ | ✅ | ✅ | ✅ | |
| drivers | ✅ | ✅ | ✅ | ✅ | |
| trucks | ✅ | ✅ | ✅ | ✅ | |
| orders | ✅ | ✅ | via admin_id | — | |
| trips | ✅ | — | ✅ | — | started_by, completed_by |
