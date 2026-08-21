# Known Issues & Improvement Notes — jb-backend

Dokumen ini mencatat keterbatasan teknis, masalah yang diketahui, dan daftar improvement yang bisa dilakukan di masa depan pada `jb-backend`.

---

## Masalah yang Diketahui (Known Issues)

### ~~1. Tidak Ada Kolom `updated_at` di Semua Tabel~~ ✅ RESOLVED
**Diselesaikan:** `migrations/init.sql` diperbarui — kolom `updated_at TIMESTAMP WITH TIME ZONE` ditambah ke tabel customers, drivers, trucks, dan orders. Semua repository UPDATE query menyetel `updated_at = CURRENT_TIMESTAMP` secara otomatis. Reset database dengan `docker-compose down -v && docker-compose up -d --build` untuk skema terbaru.

---

### ~~2. Driver Tidak Punya `created_at`~~ ✅ RESOLVED
**Diselesaikan:** `migrations/init.sql` diperbarui — kolom `created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP` ditambah ke tabel `drivers`. Model Go dan repository query sudah diperbarui.

---

### ~~3. Kolom `updated_by` Tidak Ada di customers, trucks, drivers~~ ✅ RESOLVED
**Diselesaikan:** `migrations/init.sql` diperbarui — kolom `updated_by INT REFERENCES users(id)` ditambah ke tabel customers, trucks, dan drivers. Repository UPDATE query sudah menyertakan `updated_by = $N` dari JWT user ID.

---

### 3. Tidak Ada Test Suite
**Masalah:** Tidak ada unit test maupun integration test. Semua verifikasi dilakukan secara manual via Swagger UI atau Postman collection.

**Risiko:** Refactoring atau perubahan logic bisnis sulit divalidasi tanpa test, rawan regresi.

**Solusi di masa depan:**
- Unit test untuk usecase layer (terutama state machine order/trip)
- Integration test untuk repository layer dengan database test (PostgreSQL in Docker)
- Gunakan `testing` package standar Go + `testify` untuk assertions

---

### ~~4. CORS Origins Di-hardcode di `main.go`~~ ✅ RESOLVED
**Diselesaikan:** `LoadConfig()` di `cmd/api/main.go` sudah membaca env var `CORS_ALLOWED_ORIGINS` (comma-separated). Default fallback ke `localhost:3000,localhost:3001` jika env var tidak diset.

---

### ~~5. Tidak Ada Rate Limiting~~ ✅ RESOLVED
**Diselesaikan:** Redis-based rate limiting ditambah ke dua endpoint kritis:
- **Login** (`POST /auth/login`): max 10 attempts per IP per 60 detik → 429 jika melebihi
- **GPS Location** (`POST /trips/:id/location`): throttle 30 detik per trip via Redis `SetNX` → 429 jika terlalu cepat

---

### 6. Tidak Ada Refresh Token
**Masalah:** JWT token memiliki expiry 24 jam dan tidak ada mekanisme refresh. Setelah expired, user harus login ulang dari awal.

**Solusi di masa depan:** Implementasi refresh token flow:
- Login response tambahkan `refresh_token` (expiry 7-30 hari) disimpan sebagai HttpOnly cookie
- Tambah endpoint `POST /auth/refresh` yang menerima refresh token dan mengembalikan access token baru

---

### ~~7. GPS Location: Tidak Ada Interval Throttling~~ ✅ RESOLVED
**Diselesaikan:** Termasuk dalam fix rate limiting (#5) — `location_usecase.go` kini cek Redis key `trip:{id}:location_throttle` sebelum simpan. Jika key ada (TTL 30 detik), return `ErrLocationThrottled` → handler map ke 429.

---

### 8. Migration Tidak Menggunakan Tool Resmi
**Masalah:** File migration ada di folder `migrations/` tapi tidak dijalankan via tool resmi seperti `golang-migrate` atau `goose`. Migration dijalankan manual melalui script `init.sql` saat Docker container pertama kali start.

**Dampak:** Tidak ada tracking versi migration yang ter-apply. Sulit rollback ke versi sebelumnya. Tidak ada proteksi terhadap re-run migration yang sama.

**Solusi di masa depan:** Integrasikan `golang-migrate/migrate` yang bisa dipanggil dari `main.go` saat startup, dengan tracking versi di tabel `schema_migrations`.

---

### ~~9. Password Setup Admin Hanya Sekali (No Reset Flow)~~ ✅ RESOLVED
**Diselesaikan:** Endpoint `PATCH /admin/users/:id/password` ditambah (super_admin only). Backend: `UserRepository.UpdatePasswordHash` + `UserUsecase.ResetPassword` + handler `ResetUserPassword`. Frontend: tombol "Reset Password" di halaman Users membuka modal input password baru.

---

### ~~10. `cmd/genhash/` Tidak Terdokumentasi~~ ✅ RESOLVED
**Diselesaikan:** Utility didokumentasikan di bawah.

**Cara pakai `cmd/genhash/`:**
```bash
# Generate bcrypt hash untuk password baru (misal: untuk setup manual di DB)
cd jb-backend
go run cmd/genhash/main.go

# Atau dengan Docker:
docker exec -it jbm_api go run cmd/genhash/main.go
```
Utility ini berguna jika perlu set password super_admin secara manual langsung ke PostgreSQL (tanpa melalui endpoint API).

---

### 11. Order Number dan Trip Number Generated di DB Trigger
**Masalah:** Format `ORD-{id}` dan `TRIP-{zero-padded-id}` di-generate oleh database trigger, bukan di application layer. Ini membuat logika bisnis tersembunyi di database.

**Dampak:** Sulit di-test tanpa koneksi database aktual. Format sulit diubah tanpa modifikasi trigger SQL.

**Solusi di masa depan:** Pindahkan logic generate nomor ke usecase layer Go, gunakan sequence atau UUID.

---

### ~~12. Render Crash Loop Setelah Supabase Resume~~ ✅ RESOLVED
**Gejala:** Setelah Supabase sempat di-pause lalu di-resume, service Render `jb-backend` terus mengembalikan 502 Bad Gateway. Log Supabase terlihat menerima request health/readiness berulang ke `/auth/v1/health` dan `/rest-admin/v1/ready`, karena Render terus mencoba restart service yang gagal saat startup.

**Akar masalah:** Startup dependency check terlalu agresif. `pkg/database/postgres.go` memakai `sqlx.Connect`, yang langsung melakukan `Open + Ping`, lalu `cmd/api/main.go` melakukan `db.Ping()` lagi. Redis juga langsung `Ping(ctx)` dengan startup context global hanya 10 detik. Jika Supabase pooler atau Upstash Redis butuh beberapa detik lebih lama setelah cold-start/resume, initialization gagal dan proses `os.Exit(1)`, sehingga Render masuk crash loop.

**Diselesaikan:** Startup DB dan Redis sekarang memakai retry terbatas dengan exponential backoff:
- PostgreSQL pool dibuat dengan `sqlx.Open`, lalu health check dilakukan eksplisit via `PingContext`
- Redis tetap dibuat sebagai client, lalu `Ping` diverifikasi dengan retry
- Max 8 attempts, timeout 20 detik per attempt, backoff 2 detik sampai max 30 detik
- Startup context tidak lagi fixed 10 detik, sehingga dependency punya waktu untuk siap saat cold-start

**Pembelajaran:** Untuk managed dependency seperti Supabase, Upstash, atau service serverless lain, jangan jadikan first failed ping sebagai alasan langsung crash. Startup harus membedakan antara "dependency belum siap beberapa detik" dan "dependency benar-benar down". Gunakan retry dengan backoff dan limit, logging attempt yang jelas, serta timeout per attempt yang cukup longgar.

---

## Improvement yang Sudah Dilakukan

| Tanggal | Improvement |
|---------|-------------|
| 2026-06 | `init.sql` diperbarui: `updated_at`, `updated_by`, `created_at` (drivers) ke semua tabel |
| 2026-06 | Audit trail: field `started_by_name` dan `completed_by_name` di trip |
| 2026-06 | Soft delete (`is_active`) di semua entity |
| 2026-06 | Audit trail `created_by_name` dan `updated_by_name` di semua entity |
| 2026-06 | GPS location cache di Redis + history di PostgreSQL |
| 2026-06 | Graceful shutdown (SIGINT/SIGTERM, timeout 10 detik) |
| 2026-06 | Koordinat GPS opsional di orders (origin_lat/lng, dest_lat/lng) |
| 2026-07 | Endpoint reset password admin — `PATCH /admin/users/:id/password` (super_admin only) |
| 2026-07 | Rate limiting login (10 req/IP/60s) + GPS throttle (1 req/trip/30s) via Redis |
| 2026-08 | Startup retry dengan exponential backoff untuk PostgreSQL Supabase dan Upstash Redis |

---

## Backlog Improvement

- [x] ~~Tambah kolom `updated_at` ke semua tabel~~ — selesai (`init.sql` + repository query)
- [x] ~~Tambah kolom `created_at` ke tabel `drivers`~~ — selesai (`init.sql`)
- [x] ~~Tambah kolom `updated_by` ke customers, trucks, drivers~~ — selesai (`init.sql`)
- [ ] Implementasi refresh token
- [x] ~~Rate limiting untuk endpoint login dan GPS location~~ — selesai (Redis-based, 429 response)
- [ ] Integrasi `golang-migrate` untuk versioned migration management
- [ ] Unit test untuk usecase layer
- [ ] Integration test untuk repository layer
- [x] ~~Pindahkan CORS origins ke environment variable~~ — selesai (sudah baca dari `CORS_ALLOWED_ORIGINS` env var)
- [x] ~~Endpoint reset password admin~~ — selesai (`PATCH /admin/users/:id/password`)
- [x] ~~Dokumentasi `cmd/genhash/`~~ — selesai (didokumentasikan di KNOWN_ISSUES #10)
- [ ] Pindahkan logika generate order/trip number dari DB trigger ke Go
- [x] ~~Throttling GPS location~~ — selesai (Redis SetNX 30 detik per trip)
- [x] ~~Startup retry untuk DB/Redis managed dependency~~ — selesai (`sqlx.Open` + retry `PingContext`/Redis `Ping`)
