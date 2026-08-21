# Backend Refactor Notes

Catatan ini merangkum refactor backend yang sudah diterapkan agar alasan teknis dan lokasi perubahan mudah dilacak.

---

## 2026-07-25 - Request Context Propagation

### Tujuan

Memastikan `context.Context` dari HTTP request Gin diteruskan dari Handler ke Usecase sampai Repository, sehingga cancellation signal dari client disconnect, request timeout, atau server shutdown bisa ikut menghentikan query PostgreSQL/Redis yang sedang berjalan.

### Perubahan

Sebelumnya banyak handler membuat context baru dari:

```go
context.WithTimeout(context.Background(), 5*time.Second)
```

Pola tersebut memberi timeout lokal 5 detik, tetapi tidak membawa cancellation dari request HTTP.

Sekarang handler memakai:

```go
context.WithTimeout(c.Request.Context(), 5*time.Second)
```

### File Terdampak

- `internal/handler/auth_handler.go`
- `internal/handler/customer_routes.go`
- `internal/handler/dashboard_routes.go`
- `internal/handler/driver_routes.go`
- `internal/handler/location_routes.go`
- `internal/handler/order_routes.go`
- `internal/handler/public_routes.go`
- `internal/handler/trip_routes.go`
- `internal/handler/truck_routes.go`
- `internal/handler/user_routes.go`

### Dampak Teknis

Alur context sekarang menjadi:

```text
Gin Request Context
  -> Handler context.WithTimeout(c.Request.Context(), 5*time.Second)
  -> Usecase method(ctx, ...)
  -> Repository QueryRowContext / ExecContext / GetContext / SelectContext
  -> PostgreSQL / Redis driver
```

Jika request dibatalkan, context akan ikut canceled dan operasi database yang menerima context dapat berhenti lebih cepat.

Commit terkait:

```text
8f46302 refactor: propagate request context through handlers
```

---

## 2026-07-25 - Atomic Transaction untuk Workflow Trip

### Tujuan

Mencegah data setengah tersimpan pada workflow multi-query Trip. Fokus utama refactor ini adalah alur `CreateTrip`, karena proses tersebut menyentuh beberapa tabel sekaligus: `trips`, `trucks`, `drivers`, `orders`, dan `audit_logs`.

Sebelum refactor, query dijalankan satu per satu dari usecase. Jika insert trip berhasil tetapi update status truck/driver atau audit log gagal, database bisa berada dalam kondisi tidak konsisten.

### Function Utama: CreateWithAssignments

Ditambahkan di `internal/repository/trip_repository.go`:

```go
func (r *TripRepository) CreateWithAssignments(ctx context.Context, t *model.Trip, updateOrderStatus bool, auditUserID *int) error
```

Function ini membungkus eksekusi query dalam satu database transaction (`sqlx.Tx`) menggunakan `BeginTxx(ctx, nil)` dengan mekanisme `defer tx.Rollback()`.

Alur atomic untuk `CreateTrip`:

- insert trip baru ke `trips`
- update `trucks.status = 'on_duty'`
- update `drivers.status = 'on_duty'`
- update `orders.status = 'partial'` jika order sebelumnya `pending`
- insert catatan ke `audit_logs` jika audit aktif

Jika semua query berhasil, transaction diakhiri dengan `tx.Commit()`. Jika salah satu query gagal, function return error dan `defer tx.Rollback()` membatalkan seluruh perubahan dalam transaction tersebut.

### Contoh Tambahan: CompleteWithRelease

Untuk alur penyelesaian trip, ditambahkan juga:

```go
func (r *TripRepository) CompleteWithRelease(ctx context.Context, t *model.Trip, endTime time.Time, completedBy *int, oldValues string, auditUserID *int) error
```

Function ini memakai pola transaction yang sama untuk:

- update trip menjadi `delivered`
- update truck menjadi `available`
- update driver menjadi `available`
- hitung total trip dan delivered trip untuk order terkait
- update order menjadi `completed` atau `partial`
- insert audit log jika audit aktif

### Pola Transaction

Kedua function memakai pola eksplisit berikut:

```go
tx, err := r.db.BeginTxx(ctx, nil)
if err != nil {
	return err
}
defer tx.Rollback()

// jalankan semua query memakai tx.QueryRowContext / tx.ExecContext / tx.GetContext

return tx.Commit()
```

`BeginTxx(ctx, nil)` memastikan transaction terikat ke `context.Context` yang diterima dari layer atas. Semua query di dalam transaction dijalankan lewat context-aware method seperti `tx.QueryRowContext`, `tx.ExecContext`, dan `tx.GetContext`.

Jika salah satu query gagal, function return error dan `defer tx.Rollback()` membatalkan perubahan yang sudah terjadi dalam transaction tersebut. Jika semua query berhasil, `tx.Commit()` menyimpan semua perubahan sebagai satu unit atomic.

### Hubungan dengan Pembatalan Request

Untuk menangani pembatalan koneksi dari client, handler mengalirkan `c.Request.Context()` dari Gin Handler ke Usecase hingga Repository.

Alurnya:

```text
Gin Handler
  -> context.WithTimeout(c.Request.Context(), 5*time.Second)
  -> TripUsecase.CreateTrip(ctx, ...)
  -> TripRepository.CreateWithAssignments(ctx, ...)
  -> r.db.BeginTxx(ctx, nil)
  -> tx.QueryRowContext / tx.ExecContext
```

Jika client memutus koneksi di tengah jalan atau request timeout, context akan menerima sinyal cancellation melalui `ctx.Done()`. Query database yang memakai context tersebut akan return error, lalu `defer tx.Rollback()` menjalankan rollback transaction sehingga perubahan parsial tidak tersimpan.

### Perubahan Usecase

Di `internal/usecase/trip_usecase.go`, business validation tetap berada di usecase:

- validasi ID order/truck/driver
- cek trip sudah ada atau belum
- cek order masih bisa diproses
- cek truck dan driver aktif serta `available`
- cek status trip sebelum complete

Mutation multi-query dipindahkan ke repository transaction method:

```go
u.tripRepo.CreateWithAssignments(ctx, trip, order.Status == "pending", auditUserID)
```

Untuk alur complete trip, usecase memakai transaction method tambahan:

```go
u.tripRepo.CompleteWithRelease(ctx, trip, now, &actorUserID, string(oldValue), auditUserID)
```

### Dampak Teknis

- `CreateTrip` sekarang atomic: insert trip, assign truck/driver, update order, dan audit log commit bersama atau rollback bersama.
- `CompleteTrip` sekarang atomic: delivered trip, release truck/driver, update order, dan audit log commit bersama atau rollback bersama.
- Audit log pada workflow transaction tidak lagi diabaikan jika gagal.
- HTTP API, route, model request/response, dan middleware tidak berubah.

Commit terkait:

```text
4fb4d0f refactor: wrap trip workflows in database transactions
```

---

## Verifikasi

Perintah verifikasi yang sudah dijalankan:

```bash
go test ./...
```

Hasil: semua package berhasil compile. Project belum memiliki test suite otomatis, sehingga output masih berupa `[no test files]` untuk package yang ada.
