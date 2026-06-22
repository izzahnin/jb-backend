# API Testing — Trips & Location

Base URL: `http://localhost:8080`

Semua endpoint Trips dan Location membutuhkan **role `admin_ops`** (atau `super_admin`).

---

## Autentikasi

Semua request ke endpoint Trips/Location harus menyertakan header:
```
Authorization: Bearer <token>
```

Login terlebih dahulu untuk mendapatkan token:

```http
POST /auth/login
Content-Type: application/json

{
  "username": "admin_ops",
  "password": "password123"
}
```

**Response sukses (200):**
```json
{
  "token": "eyJhbGci...",
  "expires_at": 1719000000,
  "user": {
    "id": 1,
    "username": "admin_ops",
    "role": "admin_ops"
  }
}
```

Salin nilai `token` dan gunakan di semua request berikutnya.

---

## TRIPS

### 1. List Semua Trips

```http
GET /admin/trips
Authorization: Bearer <token>
```

**Response sukses (200):**
```json
{
  "message": "Trips retrieved successfully",
  "count": 2,
  "data": [
    {
      "id": 1,
      "order_id": 1,
      "truck_id": 1,
      "driver_id": 1,
      "trip_number": "TRIP-001",
      "container_number": "",
      "seal_number": "",
      "status": "pickup",
      "is_active": true,
      "start_time": null,
      "end_time": null,
      "created_at": "2026-06-17T08:00:00Z"
    }
  ]
}
```

---

### 2. Buat Trip Baru

Prasyarat: order dengan `order_id` harus berstatus `pending`, truck dan driver harus berstatus `available`.

```http
POST /admin/trips
Authorization: Bearer <token>
Content-Type: application/json

{
  "order_id": 1,
  "truck_id": 1,
  "driver_id": 1
}
```

**Response sukses (201):**
```json
{
  "message": "Trip berhasil dibuat",
  "data": {
    "id": 1,
    "trip_number": "TRIP-001",
    "status": "pickup",
    "order_id": 1,
    "truck_id": 1,
    "driver_id": 1
  }
}
```

**Error umum:**

| Status | Pesan | Penyebab |
|--------|-------|----------|
| 400 | Format JSON tidak valid | Body request salah |
| 404 | order not found | `order_id` tidak ada |
| 404 | truck not found | `truck_id` tidak ada |
| 404 | driver not found | `driver_id` tidak ada |
| 409 | truck inactive | Truck tidak `available` |
| 409 | driver inactive | Driver tidak `available` |

---

### 3. Get Trip by ID

```http
GET /admin/trips/1
Authorization: Bearer <token>
```

**Response sukses (200):**
```json
{
  "data": {
    "id": 1,
    "trip_number": "TRIP-001",
    "order_id": 1,
    "truck_id": 1,
    "driver_id": 1,
    "container_number": "",
    "seal_number": "",
    "status": "pickup",
    "start_time": null,
    "end_time": null,
    "created_at": "2026-06-17T08:00:00Z"
  }
}
```

**Error umum:**

| Status | Pesan | Penyebab |
|--------|-------|----------|
| 400 | invalid trip ID | ID bukan angka |
| 404 | — | Trip tidak ditemukan |

---

### 4. Mulai Trip (pickup → in_transit)

Prasyarat: trip harus berstatus `pickup`. Setelah ini GPS tracking aktif.

```http
PATCH /admin/trips/1/start
Authorization: Bearer <token>
Content-Type: application/json

{
  "container_number": "ABCD1234567",
  "seal_number": "SEAL-001"
}
```

**Response sukses (200):**
```json
{
  "message": "Trip berangkat dan status menjadi in_transit"
}
```

**Error umum:**

| Status | Pesan | Penyebab |
|--------|-------|----------|
| 400 | container_number wajib diisi | Field kosong |
| 400 | seal_number wajib diisi | Field kosong |
| 404 | trip not found | Trip ID tidak ada |
| 409 | invalid status transition | Trip tidak dalam status `pickup` |

---

### 5. Selesaikan Trip (in_transit → delivered)

Prasyarat: trip harus berstatus `in_transit`. Order terkait otomatis berubah ke `completed`.

```http
PATCH /admin/trips/1/deliver
Authorization: Bearer <token>
```

> Tidak perlu body request.

**Response sukses (200):**
```json
{
  "message": "Trip selesai dan status menjadi delivered"
}
```

**Error umum:**

| Status | Pesan | Penyebab |
|--------|-------|----------|
| 404 | trip not found | Trip ID tidak ada |
| 409 | invalid status transition | Trip tidak dalam status `in_transit` |

---

## LOCATION

Endpoint location menggunakan path `/trips/:id/...` (bukan `/admin/trips`), tapi tetap membutuhkan auth `admin_ops`.

---

### 6. Kirim Lokasi GPS

Prasyarat: trip harus berstatus **`in_transit`**. Tidak bisa kirim lokasi saat `pickup` atau `delivered`.

```http
POST /trips/1/location
Authorization: Bearer <token>
Content-Type: application/json

{
  "lat": -5.147665,
  "lon": 119.432731,
  "ts": "2026-06-17T08:30:00Z"
}
```

> Field `ts` opsional — jika tidak diisi, backend menggunakan waktu saat ini.
> Format `ts`: RFC3339 (`2006-01-02T15:04:05Z`)

**Response sukses (200):**
```json
{
  "message": "Lokasi trip berhasil disimpan"
}
```

**Contoh simulasi perjalanan (kirim beberapa titik berurutan):**
```json
{ "lat": -5.147665, "lon": 119.432731, "ts": "2026-06-17T08:00:00Z" }
{ "lat": -5.200000, "lon": 119.500000, "ts": "2026-06-17T09:00:00Z" }
{ "lat": -5.300000, "lon": 119.600000, "ts": "2026-06-17T10:00:00Z" }
```

**Error umum:**

| Status | Pesan | Penyebab |
|--------|-------|----------|
| 400 | invalid trip ID | ID bukan angka |
| 400 | invalid coordinates | lat/lon di luar range valid |
| 404 | trip not found | Trip ID tidak ada |
| 409 | trip not in transit | Trip bukan status `in_transit` |

**Validasi koordinat:**
- `lat`: -90 sampai 90
- `lon`: -180 sampai 180

---

### 7. Get Lokasi Terbaru

```http
GET /trips/1/location
Authorization: Bearer <token>
```

**Response sukses (200):**
```json
{
  "data": {
    "id": 1,
    "trip_id": 1,
    "latitude": -5.147665,
    "longitude": 119.432731,
    "created_at": "2026-06-17T08:30:00Z"
  }
}
```

**Error umum:**

| Status | Pesan | Penyebab |
|--------|-------|----------|
| 404 | Lokasi trip tidak ditemukan | Belum ada data GPS untuk trip ini |

---

### 8. Get History Lokasi

```http
GET /trips/1/locations
Authorization: Bearer <token>
```

Dengan limit (default 50, maksimum 500):
```http
GET /trips/1/locations?limit=100
Authorization: Bearer <token>
```

**Response sukses (200):**
```json
{
  "data": [
    {
      "id": 3,
      "trip_id": 1,
      "latitude": -5.300000,
      "longitude": 119.600000,
      "created_at": "2026-06-17T10:00:00Z"
    },
    {
      "id": 2,
      "trip_id": 1,
      "latitude": -5.200000,
      "longitude": 119.500000,
      "created_at": "2026-06-17T09:00:00Z"
    },
    {
      "id": 1,
      "trip_id": 1,
      "latitude": -5.147665,
      "longitude": 119.432731,
      "created_at": "2026-06-17T08:00:00Z"
    }
  ]
}
```

> Data diurutkan dari **terbaru ke terlama**.

---

## Alur Testing End-to-End

Urutan yang benar untuk menguji full lifecycle trip:

```
1. Login                    → dapatkan token
2. [Pastikan ada order]     → GET /admin/orders (status harus "pending")
3. [Pastikan ada truck]     → GET /admin/trucks (status harus "available")
4. [Pastikan ada driver]    → GET /admin/drivers (status harus "available")
5. Buat trip                → POST /admin/trips
6. Cek trip                 → GET /admin/trips/1 (status: "pickup")
7. Mulai trip               → PATCH /admin/trips/1/start
8. Kirim lokasi #1          → POST /trips/1/location
9. Kirim lokasi #2          → POST /trips/1/location
10. Get lokasi terbaru      → GET /trips/1/location
11. Get history lokasi      → GET /trips/1/locations
12. Selesaikan trip         → PATCH /admin/trips/1/deliver
13. Cek trip                → GET /admin/trips/1 (status: "delivered")
14. [Cek order]             → GET /admin/orders/1 (status: "completed")
```

---

## Catatan

- **Trip number** digenerate otomatis: `TRIP-001`, `TRIP-002`, dst.
- Satu order hanya bisa punya **satu trip** (constraint `UNIQUE` di database).
- Lokasi GPS hanya bisa dikirim saat status trip **`in_transit`**.
- Setelah trip `delivered`, data lokasi history masih bisa dibaca.
- Swagger UI tersedia di: `http://localhost:8080/swagger/index.html`
