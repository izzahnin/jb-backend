# Postman API Examples

Kumpulan semua contoh request body untuk semua endpoint. Copy-paste langsung ke Postman Raw Body.

---

## Authentication Endpoints

### 1. POST /admin/setup - Initialize First Admin (One-Time Only)

**Base URL:** `http://localhost:8080`  
**Endpoint:** `POST /admin/setup`  
**Auth:** None (Public)  
**Status Code:** 201 Created

#### Option 1: With Full Name (Recommended)

**Request Body:**
```json
{
  "username": "superadmin",
  "password": "password123",
  "full_name": "System Administrator"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InN1cGVyYWRtaW4iLCJyb2xlIjoic3VwZXJfYWRtaW4iLCJleHAiOjE3MTg3MTI4MDB9...",
  "expires_at": 1718712800,
  "user": {
    "id": 1,
    "username": "superadmin",
    "full_name": "System Administrator",
    "role": "super_admin",
    "is_active": true,
    "created_at": "2026-04-18T10:00:00Z"
  }
}
```

#### Option 2: Without Full Name (Full Name will = Username)

**Request Body:**
```json
{
  "username": "superadmin",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InN1cGVyYWRtaW4iLCJyb2xlIjoic3VwZXJfYWRtaW4iLCJleHAiOjE3MTg3MTI4MDB9...",
  "expires_at": 1718712800,
  "user": {
    "id": 1,
    "username": "superadmin",
    "full_name": "superadmin",
    "role": "super_admin",
    "is_active": true,
    "created_at": "2026-04-18T10:00:00Z"
  }
}
```

**Notes:**
- Password minimal 6 karakter
- Endpoint hanya bisa dipanggil SEKALI
- Response include JWT token → save untuk request berikutnya
- Full name optional, default = username jika kosong

**Error Cases:**
```json
{
  "error": "admin user already exists"
}
```

---

### 2. POST /auth/login - Login (Get JWT Token)

**Endpoint:** `POST /auth/login`  
**Auth:** None (Public)  
**Status Code:** 200 OK

**Request Body:**
```json
{
  "username": "superadmin",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6InN1cGVyYWRtaW4iLCJyb2xlIjoic3VwZXJfYWRtaW4iLCJleHAiOjE3MTg3OTk2MDB9...",
  "expires_at": 1718799600,
  "user": {
    "id": 1,
    "username": "superadmin",
    "full_name": "System Administrator",
    "role": "super_admin",
    "is_active": true,
    "created_at": "2026-04-18T10:00:00Z"
  }
}
```

**Notes:**
- Token valid 24 jam
- Use token in `Authorization: Bearer <token>` header untuk request berikutnya
- Password case-sensitive

**Error Cases:**
```json
{
  "error": "invalid username or password"
}
```

```json
{
  "error": "user account is inactive"
}
```

---

### 3. POST /auth/logout - Logout

**Endpoint:** `POST /auth/logout`  
**Auth:** Bearer Token Required  
**Status Code:** 200 OK

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**
```json
{}
```
(Empty body)

**Response:**
```json
{
  "message": "Logout berhasil. Silahkan hapus token dari environment."
}
```

**Notes:**
- JWT adalah stateless, jadi logout hanya instruksi client untuk delete token
- Tidak ada side effect di server
- Token masih valid sampai expires_at

---

## User Management Endpoints

### 4. POST /admin/users - Create New Admin User (Super Admin Only)

**Endpoint:** `POST /admin/users`  
**Auth:** Bearer Token (super_admin role required)  
**Status Code:** 201 Created

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

#### Create admin_sales user

**Request Body:**
```json
{
  "username": "john.sales",
  "full_name": "John Smith",
  "password": "sales123456",
  "role": "admin_sales",
  "is_active": true
}
```

**Response:**
```json
{
  "message": "User berhasil dibuat",
  "data": {
    "id": 2,
    "username": "john.sales",
    "full_name": "John Smith",
    "role": "admin_sales",
    "is_active": true,
    "created_at": "2026-04-18T10:05:00Z"
  }
}
```

#### Create admin_ops user

**Request Body:**
```json
{
  "username": "jane.ops",
  "full_name": "Jane Operations",
  "password": "ops123456",
  "role": "admin_ops",
  "is_active": true
}
```

**Response:**
```json
{
  "message": "User berhasil dibuat",
  "data": {
    "id": 3,
    "username": "jane.ops",
    "full_name": "Jane Operations",
    "role": "admin_ops",
    "is_active": true,
    "created_at": "2026-04-18T10:10:00Z"
  }
}
```

#### Create another super_admin

**Request Body:**
```json
{
  "username": "admin.backup",
  "full_name": "Backup Administrator",
  "password": "admin123456",
  "role": "super_admin",
  "is_active": true
}
```

**Response:**
```json
{
  "message": "User berhasil dibuat",
  "data": {
    "id": 4,
    "username": "admin.backup",
    "full_name": "Backup Administrator",
    "role": "super_admin",
    "is_active": true,
    "created_at": "2026-04-18T10:15:00Z"
  }
}
```

#### Create with minimal data (full_name optional, defaults to username)

**Request Body:**
```json
{
  "username": "sales.temp",
  "password": "temp123456",
  "role": "admin_sales"
}
```

**Response:**
```json
{
  "message": "User berhasil dibuat",
  "data": {
    "id": 5,
    "username": "sales.temp",
    "full_name": "sales.temp",
    "role": "admin_sales",
    "is_active": true,
    "created_at": "2026-04-18T10:20:00Z"
  }
}
```

**Notes:**
- Hanya super_admin yang bisa akses endpoint ini
- Password minimal 6 karakter
- Full_name optional, defaults to username jika kosong
- is_active optional, defaults to true jika tidak dikirim
- Role harus salah satu dari: `super_admin`, `admin_sales`, `admin_ops`

**Error Cases:**
```json
{
  "error": "username already exists"
}
```

```json
{
  "error": "role must be one of: super_admin, admin_sales, admin_ops"
}
```

```json
{
  "error": "Unauthorized - super_admin role required"
}
```

---

### 5. GET /admin/users - List All Users (Super Admin Only)

**Endpoint:** `GET /admin/users`  
**Auth:** Bearer Token (super_admin role required)  
**Status Code:** 200 OK

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:** None

**Response:**
```json
{
  "message": "Users retrieved successfully",
  "count": 3,
  "data": [
    {
      "id": 1,
      "username": "superadmin",
      "full_name": "System Administrator",
      "role": "super_admin",
      "is_active": true,
      "created_at": "2026-04-18T10:00:00Z"
    },
    {
      "id": 2,
      "username": "john.sales",
      "full_name": "John Smith",
      "role": "admin_sales",
      "is_active": true,
      "created_at": "2026-04-18T10:05:00Z"
    },
    {
      "id": 3,
      "username": "jane.ops",
      "full_name": "Jane Operations",
      "role": "admin_ops",
      "is_active": true,
      "created_at": "2026-04-18T10:10:00Z"
    }
  ]
}
```

**Notes:**
- Hanya super_admin yang bisa akses endpoint ini
- List semua admin (tidak paginated)
- Respons tidak include password_hash (secure)

**Error Cases:**
```json
{
  "error": "Unauthorized - super_admin role required"
}
```

---

### 6. PATCH /admin/profile - Update Own Profile (All Admin Roles)

**Endpoint:** `PATCH /admin/profile`  
**Auth:** Bearer Token (all authenticated admin roles)  
**Status Code:** 200 OK

**Headers:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

**⭐ IMPORTANT RULE:**
- **Only include fields you want to update in the request body**
- Do NOT include empty strings ("") or null for fields you don't want to update
- ✅ Update only full_name → `{"full_name": "..."}`
- ✅ Update only password → `{"password": "..."}`
- ✅ Update both → `{"full_name": "...", "password": "..."}`
- ❌ Update only full_name but include empty password → `{"full_name": "...", "password": ""}`

#### Update only full_name

**Request Body (✅ CORRECT):**
```json
{
  "full_name": "John Smith Updated"
}
```

**❌ DO NOT DO THIS (INCORRECT):**
```json
{
  "full_name": "John Smith Updated",
  "password": ""
}
```

**Why?** 
- Jika ingin update HANYA full_name → jangan include field password sama sekali
- Jangan kosongkan password dengan `""` atau `null`
- Include field yang ingin di-update SAJA

**Response:**
```json
{
  "message": "Profile berhasil diperbarui",
  "data": {
    "id": 2,
    "username": "john.sales",
    "full_name": "John Smith Updated",
    "role": "admin_sales",
    "is_active": true,
    "created_at": "2026-04-18T10:05:00Z"
  }
}
```

#### Update only password

**Request Body (✅ CORRECT):**
```json
{
  "password": "newpassword123456"
}
```

**❌ DO NOT DO THIS (INCORRECT):**
```json
{
  "full_name": "",
  "password": "newpassword123456"
}
```

**Why?**
- Jika ingin update HANYA password → jangan include field full_name sama sekali
- Include field yang ingin di-update SAJA

**Response:**
```json
{
  "message": "Profile berhasil diperbarui",
  "data": {
    "id": 2,
    "username": "john.sales",
    "full_name": "John Smith",
    "role": "admin_sales",
    "is_active": true,
    "created_at": "2026-04-18T10:05:00Z"
  }
}
```

#### Update both full_name and password

**Request Body:**
```json
{
  "full_name": "John Smith",
  "password": "newpassword123456"
}
```

**Response:**
```json
{
  "message": "Profile berhasil diperbarui",
  "data": {
    "id": 2,
    "username": "john.sales",
    "full_name": "John Smith",
    "role": "admin_sales",
    "is_active": true,
    "created_at": "2026-04-18T10:05:00Z"
  }
}
```

**Notes:**
- Semua authenticated admin bisa update profile mereka sendiri
- **Minimal 1 field harus diisi** (full_name atau password)
- **Hanya include fields yang ingin di-update** → jangan include field lain dengan nilai kosong atau null
- Contoh: hanya update full_name → JANGAN include password field
- Contoh: hanya update password → JANGAN include full_name field
- Password minimal 6 karakter jika diisi
- user_id diambil dari JWT token (tidak bisa update user lain)
- Respons tidak include password_hash (secure)

**Error Cases:**
```json
{
  "error": "at least one field (full_name or password) must be provided"
}
```

```json
{
  "error": "password must be at least 6 characters"
}
```

```json
{
  "error": "user not found"
}
```

```json
{
  "error": "Unauthorized"
}
```

---

## Postman Setup Guide

### 1. Set Up Environment Variables

Buat environment baru dengan variables:

```
{
  "base_url": "http://localhost:8080",
  "token": ""
}
```

### 2. Auto-Save Token After Login

Di Postman tab **Tests**, tambahkan script:

```javascript
if (pm.response.code === 200 || pm.response.code === 201) {
    var jsonData = pm.response.json();
    if (jsonData.token) {
        pm.environment.set("token", jsonData.token);
    }
}
```

### 3. Use Token in Requests

Di Postman tab **Headers**, tambahkan:

```
Key: Authorization
Value: Bearer {{token}}
```

### 4. Quick Setup Flow di Postman

1. **POST /admin/setup** (get initial token)
   - Copy response token ke environment
   
2. **POST /admin/users** (create users)
   - Token automatically set dari step 1
   
3. **GET /admin/users** (verify users)
   - Token automatically set
   
4. **PATCH /admin/profile** (update profile)
   - Token automatically set

---

## Response Status Codes Reference

| Status | Meaning | Example |
|--------|---------|---------|
| 200 | OK | Login sukses, GET requests |
| 201 | Created | User/Admin berhasil dibuat |
| 400 | Bad Request | Invalid JSON, missing required field |
| 401 | Unauthorized | Missing/invalid token |
| 403 | Forbidden | Role tidak sesuai |
| 404 | Not Found | User/resource tidak ditemukan |
| 409 | Conflict | Username sudah ada |
| 422 | Unprocessable Entity | Validation error |
| 500 | Server Error | Database/system error |

---

## Common Headers

Semua request (kecuali yang marked public) membutuhkan:

```
Authorization: Bearer <your_jwt_token>
Content-Type: application/json
```

---

## Troubleshooting Postman Copy-Paste

1. **JSON Invalid Error**
   - Pastikan JSON syntax benar (no trailing commas)
   - Gunakan Postman "Beautify" button untuk auto-format

2. **Token Expired**
   - Re-login dengan POST /auth/login untuk get fresh token
   - Atau re-run POST /admin/setup jika first time

3. **403 Forbidden**
   - Check token role (GET /admin/users to verify)
   - Pastikan user memiliki role yang diperlukan

4. **411 Length Required**
   - Di Postman, pastikan Header "Content-Length" tidak di-set manual
   - Postman akan auto-set

---

## Testing Checklist

- [ ] POST /admin/setup dengan full_name
- [ ] POST /admin/setup tanpa full_name
- [ ] POST /auth/login dengan credentials
- [ ] POST /admin/users untuk create admin_sales
- [ ] POST /admin/users untuk create admin_ops
- [ ] GET /admin/users untuk list
- [ ] PATCH /admin/profile update fullname
- [ ] PATCH /admin/profile update password
- [ ] PATCH /admin/profile update both
- [ ] POST /auth/logout
