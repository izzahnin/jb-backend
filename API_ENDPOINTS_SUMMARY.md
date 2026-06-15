# API Endpoints Summary - Quick Reference

Complete list of all authentication & user management endpoints dengan status code, auth requirement, dan contoh URL.

---

## Auth & User Management - 6 Endpoints Total

| # | Method | Endpoint | Auth Required | Status | Description |
|---|--------|----------|---------------|--------|-------------|
| 1 | POST | `/auth/login` | ❌ No | 200 | Login & get JWT token (valid 24h) |
| 2 | POST | `/admin/setup` | ❌ No | 201 | Initialize first super_admin (one-time only) |
| 3 | POST | `/auth/logout` | ✅ Bearer | 200 | Logout (client: delete token) |
| 4 | POST | `/admin/users` | ✅ Bearer (super_admin) | 201 | Create new admin user (admin_sales/admin_ops) |
| 5 | GET | `/admin/users` | ✅ Bearer (super_admin) | 200 | List all users |
| 6 | PATCH | `/admin/profile` | ✅ Bearer (any admin) | 200 | Update own profile (fullname/password) |

---

## Endpoint Details

### 1️⃣ POST /auth/login
**Status:** 200 OK

Login dengan username & password untuk mendapatkan JWT token.

**cURL:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "superadmin",
    "password": "password123"
  }'
```

**Postman URL:** `http://localhost:8080/auth/login`

**Body Reference:** Lihat POSTMAN_EXAMPLES.md section "POST /auth/login"

---

### 2️⃣ POST /admin/setup
**Status:** 201 Created

Initialize first super_admin. **HANYA BISA DIJALANKAN SEKALI!**

**cURL:**
```bash
curl -X POST http://localhost:8080/admin/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "superadmin",
    "password": "password123",
    "full_name": "System Administrator"
  }'
```

**Postman URL:** `http://localhost:8080/admin/setup`

**Body Reference:** Lihat POSTMAN_EXAMPLES.md section "POST /admin/setup"

---

### 3️⃣ POST /auth/logout
**Status:** 200 OK  
**Auth:** Bearer Token Required

Logout (stateless, hanya instruksi client untuk delete token).

**cURL:**
```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Postman URL:** `http://localhost:8080/auth/logout`

**Headers:**
```
Authorization: Bearer {{token}}
```

**Body Reference:** Lihat POSTMAN_EXAMPLES.md section "POST /auth/logout"

---

### 4️⃣ POST /admin/users
**Status:** 201 Created  
**Auth:** Bearer Token + super_admin role

Create new admin user (admin_sales, admin_ops, atau super_admin lainnya).

**cURL:**
```bash
curl -X POST http://localhost:8080/admin/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "username": "john.sales",
    "full_name": "John Smith",
    "password": "sales123456",
    "role": "admin_sales",
    "is_active": true
  }'
```

**Postman URL:** `http://localhost:8080/admin/users`

**Headers:**
```
Authorization: Bearer {{token}}
Content-Type: application/json
```

**Body Reference:** Lihat POSTMAN_EXAMPLES.md section "POST /admin/users"

---

### 5️⃣ GET /admin/users
**Status:** 200 OK  
**Auth:** Bearer Token + super_admin role

List semua admin users.

**cURL:**
```bash
curl -X GET http://localhost:8080/admin/users \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Postman URL:** `http://localhost:8080/admin/users`

**Headers:**
```
Authorization: Bearer {{token}}
```

**Body Reference:** Lihat POSTMAN_EXAMPLES.md section "GET /admin/users"

---

### 6️⃣ PATCH /admin/profile
**Status:** 200 OK  
**Auth:** Bearer Token (any authenticated admin)

Update own profile (fullname dan/atau password).

**cURL - Update fullname only:**
```bash
curl -X PATCH http://localhost:8080/admin/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "full_name": "John Smith Updated"
  }'
```

**cURL - Update password only:**
```bash
curl -X PATCH http://localhost:8080/admin/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "password": "newpassword123456"
  }'
```

**Postman URL:** `http://localhost:8080/admin/profile`

**Headers:**
```
Authorization: Bearer {{token}}
Content-Type: application/json
```

**Body Reference:** Lihat POSTMAN_EXAMPLES.md section "PATCH /admin/profile"

---

## Request/Response Status Codes

### 2xx Success
| Code | Meaning | Endpoints |
|------|---------|-----------|
| 200 | OK | GET /admin/users, POST /auth/logout, PATCH /admin/profile |
| 201 | Created | POST /admin/setup, POST /admin/users |

### 4xx Client Error
| Code | Meaning | When |
|------|---------|------|
| 400 | Bad Request | Invalid JSON, missing field, validation error |
| 401 | Unauthorized | Missing/invalid JWT token, invalid credentials |
| 403 | Forbidden | Insufficient role/permission |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Username already exists |
| 422 | Unprocessable Entity | Validation error (e.g., password < 6 chars) |

### 5xx Server Error
| Code | Meaning | When |
|------|---------|------|
| 500 | Server Error | Database error, system error |

---

## Role-Based Access Control

### Who can access which endpoint?

| Endpoint | super_admin | admin_sales | admin_ops |
|----------|------------|------------|----------|
| POST /auth/login | ✅ | ✅ | ✅ |
| POST /admin/setup | ✅ | ✅ | ✅ |
| POST /auth/logout | ✅ | ✅ | ✅ |
| POST /admin/users | ✅ Only | ❌ | ❌ |
| GET /admin/users | ✅ Only | ❌ | ❌ |
| PATCH /admin/profile | ✅ (own) | ✅ (own) | ✅ (own) |

---

## Key Concepts

### JWT Token
- **Format:** `Authorization: Bearer <token>`
- **Lifespan:** 24 hours
- **Contains:** user_id, username, role, expires_at
- **Storage:** Save after login, use in all protected requests
- **Expiry:** Re-login dengan POST /auth/login untuk get fresh token

### Full Name Defaults
- **POST /admin/setup:** full_name optional → defaults to username
- **POST /admin/users:** full_name optional → defaults to username
- **PATCH /admin/profile:** Hanya update jika diisi

### Password Rules
- **Minimum 6 characters**
- **Case-sensitive**
- **Never exposed in API responses** (security)

### One-Time Endpoints
- **POST /admin/setup:** Hanya bisa dijalankan SEKALI!
  - Setelah admin pertama dibuat, endpoint permanently disabled
  - Untuk create user berikutnya → gunakan POST /admin/users

---

## Field Requirements Quick Reference

### LoginRequest (POST /auth/login)
- `username` (required): string
- `password` (required): string

### AdminSetupRequest (POST /admin/setup)
- `username` (required): string, unique
- `password` (required): string, min 6 chars
- `full_name` (optional): string → defaults to username

### CreateUserRequest (POST /admin/users)
- `username` (required): string, unique
- `password` (required): string, min 6 chars
- `full_name` (optional): string → defaults to username
- `role` (required): `super_admin` | `admin_sales` | `admin_ops`
- `is_active` (optional): boolean → defaults to true

### UpdateProfileRequest (PATCH /admin/profile)
- `full_name` (optional): string
- `password` (optional): string, min 6 chars
- **At least one field required!**
- **Important:** Only include fields you want to update. Do NOT include empty string ("") or null for fields you're not updating.
  - ✅ Update only fullname: `{"full_name": "New Name"}`
  - ✅ Update only password: `{"password": "newpass123"}`
  - ❌ Update only fullname but include empty password: `{"full_name": "Name", "password": ""}`

---

## Common Error Messages

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `invalid request format` | Invalid JSON syntax | Check JSON formatting |
| `username required` | Missing username field | Add username to request body |
| `password required` | Missing password field | Add password to request body |
| `password must be at least 6 characters` | Password < 6 chars | Use password min 6 chars |
| `invalid username or password` | Wrong credentials | Check username & password |
| `user account is inactive` | User di-deactivate | Contact admin |
| `username already exists` | Duplicate username | Use different username |
| `admin user already exists` | Setup sudah dijalankan | Use POST /auth/login |
| `role must be one of: super_admin, admin_sales, admin_ops` | Invalid role | Use valid role |
| `at least one field must be provided` | Empty body di PATCH | Add full_name atau password |
| `Unauthorized` | Missing/invalid token | Add Authorization header dengan valid token |
| `Forbidden - super_admin role required` | Insufficient permission | Use super_admin account |
| `user not found` | User tidak ada | Check user exists di GET /admin/users |

---

## Postman Collection Tips

### 1. Import Body Examples
1. Buka POSTMAN_EXAMPLES.md
2. Copy request body dari section yang diinginkan
3. Paste ke Postman tab "Body" > "Raw" > pilih "JSON"

### 2. Set Environment Variable
1. Klik "Environment" di Postman
2. Create baru dengan `base_url = http://localhost:8080`
3. Gunakan `{{base_url}}` di Postman URL bar

### 3. Auto-Save Token
Di tab "Tests" (Postman), tambahkan:
```javascript
if (pm.response.code === 200 || pm.response.code === 201) {
    var jsonData = pm.response.json();
    if (jsonData.token) {
        pm.environment.set("token", jsonData.token);
    }
}
```

### 4. Use Token
Di tab "Headers", set:
```
Key: Authorization
Value: Bearer {{token}}
```

---

## Testing Flow

```
1. POST /admin/setup
   ↓ (saves token to environment)
2. GET /admin/users
   ↓
3. POST /admin/users (create admin_sales)
   ↓
4. POST /admin/users (create admin_ops)
   ↓
5. PATCH /admin/profile (update own fullname)
   ↓
6. POST /auth/logout
```

---

## Documentation Files Reference

- **POSTMAN_EXAMPLES.md** → Lengkap request/response examples untuk all endpoints
- **ADMIN_SETUP_FLOW.md** → Detailed workflow & best practices
- **API_ENDPOINTS_SUMMARY.md** → This file, quick reference

---

## Additional Resources

- **Swagger/OpenAPI:** Access at `http://localhost:8080/swagger/docs`
- **Backend README:** `README.md` untuk setup backend
- **Postman:** Free client untuk test API → https://www.postman.com/downloads/

---

**Last Updated:** 2026-04-18  
**Endpoint Count:** 6 (Auth & User Management)  
**Role Types:** 3 (super_admin, admin_sales, admin_ops)
