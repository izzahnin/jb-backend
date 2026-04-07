# 📡 API Reference Guide

Complete documentation for all 21 endpoints in the Jalur Berlian Backend system.

---

## 📖 Table of Contents

1. [Authentication](#authentication)
2. [Admin - User Management](#admin---user-management)
3. [Admin - Fleet Management](#admin---fleet-management)
4. [Admin - Order Management](#admin---order-management)
5. [Location Tracking](#location-tracking)
6. [Public Tracking](#public-tracking)
7. [Error Handling](#error-handling)
8. [Pagination](#pagination)
9. [Authentication Headers](#authentication-headers)
10. [Complete Workflow Examples](#complete-workflow-examples)

---

## 🔐 Authentication

### POST /auth/login
Authenticate and get JWT token.

**Authorization**: Public (no token required)

**Request Body**:
```json
{
  "username": "admin",
  "password": "securepassword123"
}
```

**Response (200 OK)**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-03-17T13:30:00Z",
  "user_id": 1,
  "username": "admin"
}
```

**Response (401 Unauthorized)**:
```json
{
  "error": "invalid username or password"
}
```

**Token Details**:
- **Algorithm**: HS256
- **Expiration**: 24 hours
- **Usage**: Include in `Authorization` header as `Bearer <token>`

**Status Codes**:
- `200` - Login successful
- `401` - Invalid credentials
- `500` - Server error

---

### POST /auth/logout
Logout user (invalidates session).

**Authorization**: Required (Bearer token)

**Request Body**: Empty

**Response (200 OK)**:
```json
{
  "message": "logout successful"
}
```

**Response (401 Unauthorized)**:
```json
{
  "error": "invalid or expired token"
}
```

**Status Codes**:
- `200` - Logout successful
- `401` - Invalid token
- `500` - Server error

---

### POST /admin/setup
Initialize the first admin user in the system (one-time setup endpoint).

**Authorization**: Public (no token required - only works if no admin exists)

**Request Body**:
```json
{
  "username": "admin",
  "password": "securepassword123"
}
```

**Response (201 Created)**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1710604800,
  "user": {
    "id": 1,
    "username": "admin",
    "role": "admin",
    "is_active": true,
    "created_at": "2026-03-16T13:30:00Z"
  }
}
```

**Response (409 Conflict)**:
```json
{
  "error": "admin user already exists"
}
```

**Response (400 Bad Request)**:
```json
{
  "error": "password must be at least 6 characters"
}
```

**Status Codes**:
- `201` - Admin created successfully (token returned)
- `400` - Invalid input (username/password validation failed)
- `409` - Admin already exists (endpoint already used)
- `500` - Server error

**Notes**:
- This endpoint can only be called ONCE to bootstrap the system
- After the first admin is created, use POST /admin/users to create additional admins
- Returns JWT token for immediate login (24-hour expiry)

---

## 👥 Admin - User Management

### POST /admin/users
Create a new admin or customer user (admin-only).

**Authorization**: Required (Bearer token, admin role)

**Request Body**:
```json
{
  "username": "newadmin",
  "password": "strongpassword456",
  "role": "admin",
  "is_active": true
}
```

**Response (201 Created)**:
```json
{
  "message": "User berhasil dibuat",
  "data": {
    "id": 2,
    "username": "newadmin",
    "role": "admin",
    "is_active": true,
    "created_at": "2026-03-16T14:15:00Z"
  }
}
```

**Response (400 Bad Request)**:
```json
{
  "error": "username already exists"
}
```

**Status Codes**:
- `201` - User created
- `400` - Invalid input or duplicate username
- `401` - Unauthorized
- `409` - Username conflict
- `500` - Server error

**Notes**:
- `is_active` defaults to `true` if not provided
- `role` must be either `"admin"` or `"customer"`
- Password must be at least 6 characters
- Token with admin role is required

---

### GET /admin/users
List all users in the system (admin-only).

**Authorization**: Required (Bearer token, admin role)

**Response (200 OK)**:
```json
{
  "message": "Users retrieved successfully",
  "count": 3,
  "data": [
    {
      "id": 1,
      "username": "admin",
      "role": "admin",
      "is_active": true,
      "created_at": "2026-03-16T13:30:00Z"
    },
    {
      "id": 2,
      "username": "driver1",
      "role": "customer",
      "is_active": true,
      "created_at": "2026-03-16T14:00:00Z"
    },
    {
      "id": 3,
      "username": "admin2",
      "role": "admin",
      "is_active": true,
      "created_at": "2026-03-16T14:15:00Z"
    }
  ]
}
```

**Response (401 Unauthorized)**:
```json
{
  "error": "Unauthorized or admin role required"
}
```

**Status Codes**:
- `200` - Success
- `401` - Unauthorized or insufficient permissions
- `500` - Server error

**Notes**:
- Only accessible to admin users
- Password hashes are not exposed
- Returns users sorted by creation date (newest first)
- Use this to audit all system users and their roles

---

## 🚛 Admin - Fleet Management

### POST /admin/trucks
Register a new truck in the fleet.

**Authorization**: Required (Bearer token)

**Request Body**:
```json
{
  "plate_number": "AB1234CD",
  "driver_name": "Budi Santoso"
}
```

**Response (201 Created)**:
```json
{
  "id": 1,
  "plate_number": "AB1234CD",
  "driver_name": "Budi Santoso",
  "status": "active",
  "created_at": "2026-03-16T13:30:00Z"
}
```

**Response (400 Bad Request)**:
```json
{
  "error": "plate number must be unique"
}
```

**Status Codes**:
- `201` - Truck created
- `400` - Invalid input or duplicate plate number
- `401` - Unauthorized
- `500` - Server error

---

### GET /admin/trucks
List all trucks with pagination.

**Authorization**: Required (Bearer token)

**Query Parameters**:
- `offset` (integer, default: 0) - Number of records to skip
- `limit` (integer, default: 10, max: 100) - Records per page

**Request Example**:
```
GET /admin/trucks?offset=0&limit=10
```

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": 1,
      "plate_number": "AB1234CD",
      "driver_name": "Budi Santoso",
      "status": "active",
      "created_at": "2026-03-16T13:30:00Z"
    },
    {
      "id": 2,
      "plate_number": "AB5678EF",
      "driver_name": "Ahmad Riyadi",
      "status": "active",
      "created_at": "2026-03-16T14:00:00Z"
    }
  ],
  "count": 2,
  "total": 25,
  "offset": 0,
  "limit": 10
}
```

**Status Codes**:
- `200` - Success
- `401` - Unauthorized
- `500` - Server error

---

### GET /admin/trucks/:id
Get details of a specific truck.

**Authorization**: Required (Bearer token)

**Path Parameters**:
- `id` (integer) - Truck ID

**Response (200 OK)**:
```json
{
  "id": 1,
  "plate_number": "AB1234CD",
  "driver_name": "Budi Santoso",
  "status": "active",
  "created_at": "2026-03-16T13:30:00Z"
}
```

**Response (404 Not Found)**:
```json
{
  "error": "truck not found"
}
```

**Status Codes**:
- `200` - Success
- `401` - Unauthorized
- `404` - Truck not found
- `500` - Server error

---

### PATCH /admin/trucks/:id
Partially update truck information (PATCH semantics - only update provided fields).

**Authorization**: Required (Bearer token)

**Path Parameters**:
- `id` (integer) - Truck ID

**Request Body** (all fields optional - only include fields to update):
```json
{
  "plate_number": "AB1234XY",
  "driver_name": "Sudarnawan",
  "is_active": true
}
```

**Examples of Partial Updates**:

*Update only driver name*:
```json
{
  "driver_name": "Budi Santoso"
}
```

*Update only plate number*:
```json
{
  "plate_number": "CD5678EF"
}
```

*Deactivate truck*:
```json
{
  "is_active": false
}
```

**Response (200 OK)**:
```json
{
  "message": "Armada berhasil diupdate",
  "data": {
    "id": 1,
    "plate_number": "AB1234CD",
    "driver_name": "Sudarnawan",
    "is_active": true
  }
}
```

**Response (400 Bad Request - validation error)**:
```json
{
  "error": "plate_number cannot be empty"
}
```

**Status Codes**:
- `200` - Updated successfully (returns actual data from database)
- `400` - Invalid input or validation error
- `401` - Unauthorized (missing/invalid token)
- `404` - Truck not found
- `500` - Server error

**Notes**:
- All fields are optional - only include the fields you want to update
- `plate_number` and `driver_name` cannot be emptied (validation will fail if attempt to empty them)
- Response always returns the complete truck object from database (not just the updated fields)
- If field is not provided in request, it will not be modified

---

### DELETE /admin/trucks/:id
Deactivate/delete a truck from the fleet.

**Authorization**: Required (Bearer token)

**Path Parameters**:
- `id` (integer) - Truck ID

**Response (200 OK)**:
```json
{
  "message": "truck deleted successfully"
}
```

**Response (404 Not Found)**:
```json
{
  "error": "truck not found"
}
```

**Status Codes**:
- `200` - Deleted successfully
- `401` - Unauthorized
- `404` - Truck not found
- `500` - Server error

---

## 📦 Admin - Order Management

### POST /admin/orders
Create a new order.

**Authorization**: Required (Bearer token)

**Request Body**:
```json
{
  "order_number": "ORD001",
  "origin": "Makassar",
  "destination": "Jakarta",
  "description": "Elektronik 50 unit",
  "quantity": 50
}
```

**Response (201 Created)**:
```json
{
  "id": 1,
  "order_number": "ORD001",
  "origin": "Makassar",
  "destination": "Jakarta",
  "description": "Elektronik 50 unit",
  "quantity": 50,
  "status": "pending",
  "truck_id": null,
  "created_at": "2026-03-16T13:30:00Z"
}
```

**Response (400 Bad Request)**:
```json
{
  "error": "order number already exists"
}
```

**Status Codes**:
- `201` - Order created
- `400` - Invalid input or duplicate order number
- `401` - Unauthorized
- `500` - Server error

---

### GET /admin/orders
List all orders with pagination.

**Authorization**: Required (Bearer token)

**Query Parameters**:
- `offset` (integer, default: 0) - Number of records to skip
- `limit` (integer, default: 10, max: 100) - Records per page
- `status` (string, optional) - Filter by status: pending, pickup, in_transit, delivered

**Request Example**:
```
GET /admin/orders?offset=0&limit=10&status=pending
```

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": 1,
      "order_number": "ORD001",
      "origin": "Makassar",
      "destination": "Jakarta",
      "description": "Elektronik 50 unit",
      "quantity": 50,
      "status": "pending",
      "truck_id": null,
      "created_at": "2026-03-16T13:30:00Z"
    }
  ],
  "count": 1,
  "total": 47,
  "offset": 0,
  "limit": 10
}
```

**Status Codes**:
- `200` - Success
- `401` - Unauthorized
- `500` - Server error

---

### GET /admin/orders/:id
Get details of a specific order.

**Authorization**: Required (Bearer token)

**Path Parameters**:
- `id` (integer) - Order ID

**Response (200 OK)**:
```json
{
  "id": 1,
  "order_number": "ORD001",
  "origin": "Makassar",
  "destination": "Jakarta",
  "description": "Elektronik 50 unit",
  "quantity": 50,
  "status": "pending",
  "truck_id": null,
  "truck": null,
  "created_at": "2026-03-16T13:30:00Z"
}
```

**Status Codes**:
- `200` - Success
- `401` - Unauthorized
- `404` - Order not found
- `500` - Server error

---

### PUT /admin/orders/:id
Update order information (allowed only if status is pending).

**Authorization**: Required (Bearer token)

**Path Parameters**:
- `id` (integer) - Order ID

**Request Body** (all fields optional):
```json
{
  "origin": "Makassar",
  "destination": "Bandung",
  "quantity": 60
}
```

**Response (200 OK)**:
```json
{
  "id": 1,
  "order_number": "ORD001",
  "origin": "Makassar",
  "destination": "Bandung",
  "quantity": 60,
  "status": "pending",
  "updated_at": "2026-03-16T14:20:00Z"
}
```

**Status Codes**:
- `200` - Updated successfully
- `400` - Invalid state or input
- `401` - Unauthorized
- `404` - Order not found
- `500` - Server error

---

### DELETE /admin/orders/:id
Cancel an order (allowed only if status is pending).

**Authorization**: Required (Bearer token)

**Path Parameters**:
- `id` (integer) - Order ID

**Response (200 OK)**:
```json
{
  "message": "order cancelled successfully"
}
```

**Response (400 Bad Request)**:
```json
{
  "error": "cannot delete order that is not pending"
}
```

**Status Codes**:
- `200` - Deleted successfully
- `400` - Invalid state (order not pending)
- `401` - Unauthorized
- `404` - Order not found
- `500` - Server error

---

### POST /admin/orders/:id/assign
Assign a truck to an order.

**Authorization**: Required (Bearer token)

**Path Parameters**:
- `id` (integer) - Order ID

**Request Body**:
```json
{
  "truck_id": 1
}
```

**Response (200 OK)**:
```json
{
  "id": 1,
  "order_number": "ORD001",
  "origin": "Makassar",
  "destination": "Jakarta",
  "status": "pending",
  "truck_id": 1,
  "truck": {
    "id": 1,
    "plate_number": "AB1234CD",
    "driver_name": "Budi Santoso"
  },
  "updated_at": "2026-03-16T14:30:00Z"
}
```

**Response (400 Bad Request)**:
```json
{
  "error": "truck not found or order already assigned"
}
```

**Status Codes**:
- `200` - Successfully assigned
- `400` - Invalid state or truck not found
- `401` - Unauthorized
- `404` - Order not found
- `500` - Server error

---

### PUT /admin/orders/:id/confirm-pickup
Confirm pickup and transition order from pending → in_pickup state.

**Authorization**: Required (Bearer token)

**Path Parameters**:
- `id` (integer) - Order ID

**Request Body**: Empty

**Response (200 OK)**:
```json
{
  "id": 1,
  "order_number": "ORD001",
  "status": "pickup",
  "truck_id": 1,
  "updated_at": "2026-03-16T15:00:00Z"
}
```

**Response (400 Bad Request)**:
```json
{
  "error": "cannot confirm pickup - order not in pending status"
}
```

**Status Codes**:
- `200` - Pickup confirmed
- `400` - Invalid state transition
- `401` - Unauthorized
- `404` - Order not found
- `500` - Server error

---

### PUT /admin/orders/:id/confirm-delivery
Confirm delivery and transition order from in_transit → delivered state.

**Authorization**: Required (Bearer token)

**Path Parameters**:
- `id` (integer) - Order ID

**Request Body**: Empty

**Response (200 OK)**:
```json
{
  "id": 1,
  "order_number": "ORD001",
  "status": "delivered",
  "truck_id": 1,
  "updated_at": "2026-03-16T16:30:00Z"
}
```

**Response (400 Bad Request)**:
```json
{
  "error": "cannot confirm delivery - order not in in_transit status"
}
```

**Status Codes**:
- `200` - Delivery confirmed
- `400` - Invalid state transition
- `401` - Unauthorized
- `404` - Order not found
- `500` - Server error

---

## 📍 Location Tracking

### POST /locations
Record GPS coordinates for a truck.

**Authorization**: Public (no token required)

**Request Body**:
```json
{
  "truck_id": 1,
  "latitude": -8.6753,
  "longitude": 120.4317,
  "timestamp": "2026-03-16T15:45:30Z"
}
```

**Response (201 Created)**:
```json
{
  "id": 123,
  "truck_id": 1,
  "latitude": -8.6753,
  "longitude": 120.4317,
  "recorded_at": "2026-03-16T15:45:30Z"
}
```

**Response (400 Bad Request)**:
```json
{
  "error": "invalid coordinates or truck_id"
}
```

**Status Codes**:
- `201` - Location recorded
- `400` - Invalid input
- `500` - Server error

---

### GET /locations/:truck_id/latest
Get the latest location of a truck.

**Authorization**: Public (no token required)

**Path Parameters**:
- `truck_id` (integer) - Truck ID

**Response (200 OK)**:
```json
{
  "truck_id": 1,
  "latitude": -8.6753,
  "longitude": 120.4317,
  "recorded_at": "2026-03-16T15:45:30Z"
}
```

**Response (404 Not Found)**:
```json
{
  "error": "no location data found for this truck"
}
```

**Status Codes**:
- `200` - Success
- `404` - No location data
- `500` - Server error

---

### GET /locations/:truck_id/history
Get location history with pagination.

**Authorization**: Public (no token required)

**Path Parameters**:
- `truck_id` (integer) - Truck ID

**Query Parameters**:
- `offset` (integer, default: 0) - Number of records to skip
- `limit` (integer, default: 50, max: 250) - Records per page

**Request Example**:
```
GET /locations/1/history?offset=0&limit=50
```

**Response (200 OK)**:
```json
{
  "truck_id": 1,
  "data": [
    {
      "id": 123,
      "latitude": -8.6753,
      "longitude": 120.4317,
      "recorded_at": "2026-03-16T15:45:30Z"
    },
    {
      "id": 122,
      "latitude": -8.6752,
      "longitude": 120.4315,
      "recorded_at": "2026-03-16T15:43:15Z"
    }
  ],
  "count": 2,
  "total": 1250,
  "offset": 0,
  "limit": 50
}
```

**Status Codes**:
- `200` - Success
- `404` - Truck not found
- `500` - Server error

---

## 🌐 Public Tracking

### GET /orders/:order_number/track
Track an order by order number (customer-facing, no authentication required).

**Authorization**: Public (no token required)

**Path Parameters**:
- `order_number` (string) - Order number (e.g., "ORD001")

**Response (200 OK)**:
```json
{
  "order": {
    "id": 1,
    "order_number": "ORD001",
    "origin": "Makassar",
    "destination": "Jakarta",
    "status": "in_transit",
    "created_at": "2026-03-16T13:30:00Z"
  },
  "truck": {
    "id": 1,
    "plate_number": "AB1234CD",
    "driver_name": "Budi Santoso"
  },
  "last_location": {
    "latitude": -8.6753,
    "longitude": 120.4317,
    "recorded_at": "2026-03-16T15:45:30Z"
  }
}
```

**Response (404 Not Found)**:
```json
{
  "error": "order not found"
}
```

**Status Codes**:
- `200` - Success
- `404` - Order not found
- `500` - Server error

---

## ❌ Error Handling

All errors return consistent JSON format:

```json
{
  "error": "error message",
  "code": "ERROR_CODE"
}
```

### Common Error Codes

| Code | HTTP Status | Meaning |
|------|-------------|---------|
| `INVALID_INPUT` | 400 | Request body failed validation |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource already exists (duplicate) |
| `INVALID_STATE` | 400 | Invalid state transition |
| `SERVER_ERROR` | 500 | Internal server error |

---

## 📄 Pagination

List endpoints support pagination with consistent format:

**Request Parameters**:
```
?offset=0&limit=10
```

**Response**:
```json
{
  "data": [...],      // Array of resources
  "count": 10,        // Items in current page
  "total": 147,       // Total items available
  "offset": 0,        // Current offset
  "limit": 10         // Items per page
}
```

**Defaults & Limits**:
- Default offset: 0
- Default limit: 10 (trucks/orders) or 50 (locations)
- Maximum limit: 100 (trucks/orders) or 250 (locations)

---

## 🔑 Authentication Headers

### Bearer Token Format

```
Authorization: Bearer <token>
```

**Example**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkFkbWluIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

### Getting a Token

**First Time Setup (bootstrap system)**:
1. Call `POST /admin/setup` to create the first admin:
   ```bash
   POST /admin/setup
   {
     "username": "admin",
     "password": "password123"
   }
   ```
   Response includes JWT token for immediate login

**Subsequent Logins**:
1. Use `POST /auth/login` with credentials:
   ```bash
   POST /auth/login
   {
     "username": "admin",
     "password": "password123"
   }
   ```

3. Extract token from response and use in all protected endpoints

---

## 🔄 Complete Workflow Examples

### Example 1: Complete Order Lifecycle

```bash
# 1. Setup Admin (FIRST TIME ONLY - bootstrap system)
POST /admin/setup
{
  "username": "admin",
  "password": "pass123"
}
# Response: token = abc123xyz, user role = admin

# 2. Login (subsequent times)
POST /auth/login
{
  "username": "admin",
  "password": "pass123"
}
# Response: token = abc123xyz

# 3. List all users (admin audit)
GET /admin/users
Headers: Authorization: Bearer abc123xyz
# Response: All users with username, role, is_active

# 4. Create another admin or driver
POST /admin/users
Headers: Authorization: Bearer abc123xyz
{
  "username": "driver1",
  "password": "pass123",
  "role": "customer",
  "is_active": true
}
# Response: user_id = 2

# 5. Create Truck
POST /admin/trucks
Headers: Authorization: Bearer abc123xyz
{
  "plate_number": "AB1234CD",
  "driver_name": "Budi Santoso"
}
# Response: truck_id = 1

# 6. Create Order
POST /admin/orders
Headers: Authorization: Bearer abc123xyz
{
  "order_number": "ORD001",
  "origin": "Makassar",
  "destination": "Jakarta"
}
# Response: order_id = 1, status = pending

# 7. Assign Truck to Order
POST /admin/orders/1/assign
Headers: Authorization: Bearer abc123xyz
{
  "truck_id": 1
}
# Response: order status still pending, truck assigned

# 8. Confirm Pickup
PUT /admin/orders/1/confirm-pickup
Headers: Authorization: Bearer abc123xyz
# Response: status = pickup

# 9. Record Location (by IoT device, public)
POST /locations
{
  "truck_id": 1,
  "latitude": -8.6753,
  "longitude": 120.4317
}

# 10. Confirm Delivery
PUT /admin/orders/1/confirm-delivery
Headers: Authorization: Bearer abc123xyz
# Response: status = delivered

# 11. Customer Tracks Order (public)
GET /orders/ORD001/track
# Response: Complete order details with latest location
```

### Example 2: Real-Time Tracking Flow

```bash
# 1. Truck sends locations every 5 seconds (IoT device, public API)
POST /locations
{
  "truck_id": 1,
  "latitude": -8.6753,
  "longitude": 120.4317
}

# 2. Customer checks order status
GET /orders/ORD001/track
# Gets latest coordinates from Redis cache

# 3. Admin monitors fleet
GET /admin/trucks - see all trucks
GET /admin/orders - see all assignments
```

---

## 📊 Request/Response Size Limits

- **Maximum request body**: 10 MB
- **Maximum query string**: 4 KB
- **Connection timeout**: 30 seconds
- **Read timeout**: 30 seconds

---

Last Updated: 2026-03-30 | Version: 1.2.0

**Changes in v1.2.0**:
- Removed `POST /auth/register` endpoint (not needed for this project)
- Kept `POST /admin/setup` - One-time admin initialization (bootstrap)
- Kept `/auth/login` & `/auth/logout` - Admin authentication
- Kept `POST /admin/users` - Create additional admins/users
- Kept `GET /admin/users` - Audit all users and their roles
- Optional Users Auth: Only admin needs login via /auth/login after initial setup
- Public Users: No authentication needed for `GET /orders/:order_number/track`

For setup instructions, see [SETUP.md](SETUP.md)  
For development guide, see [DEVELOPMENT.md](DEVELOPMENT.md)
