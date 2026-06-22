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

### 4. CORS Origins Di-hardcode di `main.go`
**Masalah:** Daftar origins yang diizinkan CORS (`localhost:3000`, `3001`, `5173`, dst) di-hardcode langsung di `cmd/api/main.go`, bukan dibaca dari environment variable.

**Dampak:** Saat deployment ke staging/production, harus ubah kode dan rebuild.

**Solusi di masa depan:** Tambah `CORS_ALLOWED_ORIGINS` di env var, parse sebagai comma-separated string di `LoadConfig()`.

---

### 5. Tidak Ada Rate Limiting
**Masalah:** Tidak ada pembatasan jumlah request per IP/user. Endpoint login dan POST location GPS rentan terhadap brute force atau spam request.

**Solusi di masa depan:** Implementasi rate limiting menggunakan `golang.org/x/time/rate` atau middleware Gin pihak ketiga (`gin-contrib/ratelimit`).

---

### 6. Tidak Ada Refresh Token
**Masalah:** JWT token memiliki expiry 24 jam dan tidak ada mekanisme refresh. Setelah expired, user harus login ulang dari awal.

**Solusi di masa depan:** Implementasi refresh token flow:
- Login response tambahkan `refresh_token` (expiry 7-30 hari) disimpan sebagai HttpOnly cookie
- Tambah endpoint `POST /auth/refresh` yang menerima refresh token dan mengembalikan access token baru

---

### 7. GPS Location: Tidak Ada Interval Throttling
**Masalah:** Endpoint `POST /trips/{id}/location` tidak membatasi frekuensi pengiriman lokasi. Client bisa spam request setiap detik atau lebih cepat, yang bisa menyebabkan database tumbuh sangat cepat.

**Solusi di masa depan:** Tambah validasi di `location_usecase.go` — tolak request jika lokasi terbaru sudah dikirim kurang dari X detik yang lalu (cek dari Redis cache).

---

### 8. Migration Tidak Menggunakan Tool Resmi
**Masalah:** File migration ada di folder `migrations/` tapi tidak dijalankan via tool resmi seperti `golang-migrate` atau `goose`. Migration dijalankan manual melalui script `init.sql` saat Docker container pertama kali start.

**Dampak:** Tidak ada tracking versi migration yang ter-apply. Sulit rollback ke versi sebelumnya. Tidak ada proteksi terhadap re-run migration yang sama.

**Solusi di masa depan:** Integrasikan `golang-migrate/migrate` yang bisa dipanggil dari `main.go` saat startup, dengan tracking versi di tabel `schema_migrations`.

---

### 9. Password Setup Admin Hanya Sekali (No Reset Flow)
**Masalah:** Tidak ada endpoint untuk reset password admin. Jika admin lupa password, satu-satunya cara adalah generate hash baru via `cmd/genhash/` dan update langsung ke database.

**Solusi di masa depan:** Tambah endpoint `PATCH /admin/users/{id}/password` yang hanya bisa diakses `super_admin`, untuk reset password admin lain.

---

### 10. `cmd/genhash/` Tidak Terdokumentasi
**Masalah:** Utility `cmd/genhash/main.go` untuk generate bcrypt hash password ada di codebase tapi tidak disebutkan di README atau SETUP.md.

**Penanganan saat ini:** Hanya diketahui dari membaca kode langsung.

**Solusi di masa depan:** Tambahkan section di SETUP.md tentang cara menggunakan genhash untuk setup password awal.

---

### 11. Order Number dan Trip Number Generated di DB Trigger
**Masalah:** Format `ORD-{id}` dan `TRIP-{zero-padded-id}` di-generate oleh database trigger, bukan di application layer. Ini membuat logika bisnis tersembunyi di database.

**Dampak:** Sulit di-test tanpa koneksi database aktual. Format sulit diubah tanpa modifikasi trigger SQL.

**Solusi di masa depan:** Pindahkan logic generate nomor ke usecase layer Go, gunakan sequence atau UUID.

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

---

## Backlog Improvement

- [x] ~~Tambah kolom `updated_at` ke semua tabel~~ — selesai (`init.sql` + repository query)
- [x] ~~Tambah kolom `created_at` ke tabel `drivers`~~ — selesai (`init.sql`)
- [x] ~~Tambah kolom `updated_by` ke customers, trucks, drivers~~ — selesai (`init.sql`)
- [ ] Implementasi refresh token
- [ ] Rate limiting untuk endpoint login dan GPS location
- [ ] Integrasi `golang-migrate` untuk versioned migration management
- [ ] Unit test untuk usecase layer
- [ ] Integration test untuk repository layer
- [ ] Pindahkan CORS origins ke environment variable
- [ ] Endpoint reset password admin
- [ ] Dokumentasi `cmd/genhash/` di SETUP.md
- [ ] Pindahkan logika generate order/trip number dari DB trigger ke Go
- [ ] Throttling GPS location (tolak jika interval terlalu pendek)
