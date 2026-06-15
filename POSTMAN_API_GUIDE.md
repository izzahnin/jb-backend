# Postman API Guide

Base URL: `http://localhost:8080`

## Headers
- `Content-Type: application/json` for requests with body
- `Authorization: Bearer {{token}}` for protected admin routes

## Authentication

### `POST /admin/setup`
One-time create first super admin.
```json
{
  "username": "superadmin",
  "password": "password123",
  "full_name": "System Administrator"
}
```

### `POST /auth/login`
```json
{
  "username": "superadmin",
  "password": "password123"
}
```

### `POST /auth/logout`
No body.

## Dashboard

### `GET /admin/dashboard/stats`
No body.

### `PATCH /admin/profile`
At least one field is required.
```json
{
  "full_name": "Bambang Suryanto",
  "password": "newpassword123"
}
```

## Users

### `GET /admin/users`
No body.

### `POST /admin/users`
```json
{
  "username": "ops_001",
  "full_name": "Ops Admin",
  "password": "password123",
  "role": "admin_ops",
  "is_active": true
}
```

## Customers

### `GET /admin/customers`
No body.

### `POST /admin/customers`
```json
{
  "company_name": "PT. Nusantara Logistik",
  "pic_name": "Budi Santoso",
  "phone": "+628123456789",
  "email": "budi@nusantara.co.id",
  "address": "Jl. Sudirman No. 123, Jakarta",
  "npwp": "01.234.567.8-901.000"
}
```

### `GET /admin/customers/:id`
No body.

### `PATCH /admin/customers/:id`
Send only fields you want to change.
```json
{
  "company_name": "PT. Nusantara Logistik Baru",
  "phone": "+628123456780"
}
```

### `DELETE /admin/customers/:id`
No body.

## Drivers

### `GET /admin/drivers`
No body.

### `POST /admin/drivers`
```json
{
  "name": "Andi Wijaya",
  "license_number": "SIM-B-1234567",
  "phone": "+6281122334455",
  "status": "available",
  "is_active": true
}
```

### `GET /admin/drivers/:id`
No body.

### `PATCH /admin/drivers/:id`
Send only fields you want to change.
```json
{
  "status": "off",
  "phone": "+6281122334499"
}
```

### `DELETE /admin/drivers/:id`
No body.

## Trucks

### `GET /admin/trucks`
No body. Supports `limit` and `offset` query params.

### `POST /admin/trucks`
```json
{
  "plate_number": "B-1234-XYZ",
  "truck_type": "Fuso Box",
  "status": "available",
  "is_active": true
}
```

### `GET /admin/trucks/:id`
No body.

### `PATCH /admin/trucks/:id`
Send only fields you want to change.
```json
{
  "status": "maintenance",
  "truck_type": "Fuso Double"
}
```

### `DELETE /admin/trucks/:id`
No body.

## Orders

### `GET /admin/orders`
No body. Supports `limit` and `offset` query params.

### `POST /admin/orders`
`order_number` is generated automatically by the backend. Do not send it.
`total_containers` must be `1`.
```json
{ 
  "customer_id": 1,
  "origin": "Jakarta",
  "destination": "Surabaya",
  "total_containers": 1
}
```

### `GET /admin/orders/:id`
No body.

### `PATCH /admin/orders/:id`
```json
{
  "status": "partial"
}
```

### `DELETE /admin/orders/:id`
No body.

## Trips

### `GET /admin/trips`
No body.

### `POST /admin/trips`
`trip_number` is generated automatically by the backend. Do not send it.
Only one trip is allowed per order.
```json
{
  "order_id": 1,
  "truck_id": 1,
  "driver_id": 1
}
```

### `GET /admin/trips/:id`
No body. Returns one trip object by trip ID.

### `PATCH /admin/trips/:id/start`
```json
{
  "container_number": "CONT-123456",
  "seal_number": "SEAL-789012"
}
```

### `PATCH /admin/trips/:id/deliver`
No body.

## Locations

### `POST /trips/:id/location`
Only allowed when the trip status is `in_transit`.
```json
{
  "lat": -6.200000,
  "lon": 106.816666,
  "ts": "2026-04-19T08:30:00Z"
}
```

### `GET /trips/:id/location`
No body.

### `GET /trips/:id/locations?limit=50`
No body.

## Public Tracking

### `GET /public/orders/:order_number/track`
No body.
Returns the order and one trip only.

## Recommended Postman Flow
1. `POST /admin/setup` or `POST /auth/login` for token.
2. Save token into `{{token}}`.
3. Create master data in this order: customers, drivers, trucks.
4. Create an order, then create a trip.
5. Start trip, post location, complete trip.
6. Test public tracking with the returned `order_number`.
