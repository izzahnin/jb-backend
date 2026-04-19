# API Testing & Setup Checklist

Panduan lengkap untuk setup, testing, dan troubleshooting endpoints auth & user management.

---

## 🚀 Quick Start (5 Minutes)

### Step 1: Start Backend
```bash
cd jb-backend
docker-compose up -d postgres redis
go run cmd/api/main.go
```

Wait for: `Server running on :8080`

### Step 2: Setup Admin (First Time Only)

In Postman, POST to `http://localhost:8080/admin/setup`:

**Body:**
```json
{
  "username": "superadmin",
  "password": "password123",
  "full_name": "System Administrator"
}
```

**✅ Save the token from response!**

### Step 3: Verify Setup

In Postman, GET to `http://localhost:8080/admin/users`:

**Headers:**
```
Authorization: Bearer <your_token_here>
```

You should see the admin user you just created.

**Done! ✅**

---

## 📋 Full Testing Checklist

### Authentication Flow

- [ ] **Step 1: POST /admin/setup**
  - Body: username, password, full_name (optional)
  - Expected: 201 Created
  - Check: Response includes token & user object
  - Save: Token to Postman environment

- [ ] **Step 2: POST /auth/login**
  - Body: superadmin username & password
  - Expected: 200 OK
  - Check: Token matches setup token or different
  - Note: Token should be valid for 24 hours

- [ ] **Step 3: GET /admin/users** (with token)
  - Expected: 200 OK
  - Check: Response includes admin user created in step 1
  - Check: count = 1, data array has 1 user

- [ ] **Step 4: POST /auth/logout** (with token)
  - Expected: 200 OK
  - Check: Response includes message
  - Note: Token is still valid (stateless), logout happens on client

### User Management Flow

- [ ] **Step 5: POST /admin/users** (create admin_sales)
  - Token: super_admin token
  - Body: username, password, full_name, role="admin_sales"
  - Expected: 201 Created
  - Check: User ID incremented (id > 1)

- [ ] **Step 6: POST /admin/users** (create admin_ops)
  - Token: super_admin token
  - Body: username, password, full_name, role="admin_ops"
  - Expected: 201 Created
  - Check: Different user created

- [ ] **Step 7: POST /admin/users** (create another super_admin)
  - Token: super_admin token
  - Body: username, password, full_name, role="super_admin"
  - Expected: 201 Created
  - Check: Can create multiple super_admins

- [ ] **Step 8: GET /admin/users** (verify all users)
  - Token: super_admin token
  - Expected: 200 OK
  - Check: count = 4 (or however many created)
  - Check: All users have correct roles

- [ ] **Step 9: POST /admin/users with minimal data** (no full_name)
  - Token: super_admin token
  - Body: username, password, role (full_name omitted)
  - Expected: 201 Created
  - Check: Full_name defaults to username in response

### Profile Update Flow

- [ ] **Step 10: PATCH /admin/profile** (update fullname only)
  - Token: super_admin token
  - Body: {"full_name": "Updated Name"}
  - Expected: 200 OK
  - Check: Full name updated, username unchanged

- [ ] **Step 11: PATCH /admin/profile** (update password only)
  - Token: super_admin token
  - Body: {"password": "newpassword123"}
  - Expected: 200 OK
  - Check: Can login with new password

- [ ] **Step 12: PATCH /admin/profile** (update both)
  - Token: super_admin token
  - Body: {"full_name": "New Name", "password": "another123"}
  - Expected: 200 OK
  - Check: Both updated

- [ ] **Step 13: Login with new password**
  - POST /auth/login with updated password
  - Expected: 200 OK
  - Check: Fresh token generated

### Error Handling

- [ ] **POST /admin/setup** (second time)
  - Expected: 409 Conflict
  - Check: Error message = "admin user already exists"

- [ ] **POST /admin/users** (duplicate username)
  - Body: username that already exists
  - Expected: 409 Conflict
  - Check: Error message = "username already exists"

- [ ] **POST /admin/users** (invalid role)
  - Body: role="invalid_role"
  - Expected: 400 Bad Request
  - Check: Error about valid roles

- [ ] **POST /admin/users** (no token)
  - No Authorization header
  - Expected: 401 Unauthorized
  - Check: Error about missing auth

- [ ] **POST /admin/users** (non-super_admin token)
  - Token: admin_sales token
  - Expected: 403 Forbidden
  - Check: Error about super_admin role required

- [ ] **POST /admin/users** (invalid json)
  - Body: {invalid json}
  - Expected: 400 Bad Request
  - Check: Error about JSON format

- [ ] **PATCH /admin/profile** (no fields)
  - Body: {}
  - Expected: 400 Bad Request
  - Check: Error about at least one field required

- [ ] **PATCH /admin/profile** (password too short)
  - Body: {"password": "short"}
  - Expected: 400 Bad Request
  - Check: Error about password min 6 chars

---

## 🔧 Troubleshooting Guide

### Problem: "admin user already exists" on setup

**Cause:** Setup sudah pernah dijalankan  
**Solution:**
1. Check database sudah ada super_admin user
2. Use POST /auth/login untuk login
3. Jika lupa password, perlu reset database:
   ```bash
   docker-compose down -v
   docker-compose up -d postgres redis
   ```

### Problem: "Unauthorized" on protected endpoint

**Cause:** Missing atau invalid token  
**Solution:**
1. Check Authorization header present: `Authorization: Bearer <token>`
2. Check token tidak expired (24 hours dari login)
3. Re-login: POST /auth/login untuk get fresh token
4. Check token format (copy tanpa quotes)

### Problem: "Forbidden - super_admin role required"

**Cause:** Token dari admin_sales/admin_ops, bukan super_admin  
**Solution:**
1. Use token dari super_admin user
2. Check GET /admin/users untuk verify user role
3. If wrong user, create new super_admin dengan POST /admin/users

### Problem: "username already exists"

**Cause:** Username sudah dipakai  
**Solution:**
1. Use different username
2. Or check apakah user sudah ada di GET /admin/users

### Problem: "password must be at least 6 characters"

**Cause:** Password field < 6 karakter  
**Solution:**
1. Use password minimal 6 karakter
2. Example: "pass123", "mypassword2024"

### Problem: 500 Internal Server Error

**Cause:** Database connection error  
**Solution:**
1. Check Docker containers: `docker-compose ps`
2. Verify postgres & redis running (status = healthy)
3. Check connection string di .env
4. Restart containers: `docker-compose restart`

### Problem: "Format JSON tidak valid"

**Cause:** JSON body syntax error  
**Solution:**
1. Validate JSON format (use online validator)
2. Check no trailing commas
3. Check quotes around keys & string values
4. Postman: click "Beautify" button untuk auto-format

---

## 📊 Test Data Reference

### Predefined Test Users

Create these users untuk comprehensive testing:

#### User 1: Sales Team Lead
```json
{
  "username": "sales.lead",
  "full_name": "Budi Santoso",
  "password": "saleslead123",
  "role": "admin_sales"
}
```

#### User 2: Operations Manager
```json
{
  "username": "ops.manager",
  "full_name": "Siti Nurhaliza",
  "password": "opsmanager123",
  "role": "admin_ops"
}
```

#### User 3: Backup Admin
```json
{
  "username": "admin.backup",
  "full_name": "Ahmad Wijaya",
  "password": "backup123456",
  "role": "super_admin"
}
```

---

## 📁 File References

| File | Purpose | When to Use |
|------|---------|------------|
| **POSTMAN_EXAMPLES.md** | All request/response examples | Copy-paste to Postman |
| **API_ENDPOINTS_SUMMARY.md** | Quick reference all endpoints | Overview & URLs |
| **ADMIN_SETUP_FLOW.md** | Detailed workflow & best practices | Understanding flow |
| **API_TESTING_CHECKLIST.md** | This file | Step-by-step testing |

---

## 🔐 Security Checklist

Before production deployment:

- [ ] Change default passwords
- [ ] Set strong JWT secret in .env
- [ ] Enable HTTPS (not just HTTP)
- [ ] Rate limiting on auth endpoints
- [ ] Password complexity requirements
- [ ] Session timeout < 24 hours
- [ ] Implement refresh token rotation
- [ ] Log all auth attempts
- [ ] Regular security audit
- [ ] Backup strategy for database

---

## 🎯 Performance Testing

### Load Test Setup (optional)

```bash
# Using Apache Bench (example: 100 requests, 10 concurrent)
ab -n 100 -c 10 http://localhost:8080/auth/login -p payload.json
```

### Expected Response Times
- POST /auth/login: < 100ms
- GET /admin/users: < 50ms
- POST /admin/users: < 100ms
- PATCH /admin/profile: < 100ms

---

## 📝 Postman Setup Recommendations

### 1. Create Collections
- Collection: "Jalur Berlian Backend"
  - Folder: "Auth"
    - Request: POST /auth/login
    - Request: POST /admin/setup
    - Request: POST /auth/logout
  - Folder: "User Management"
    - Request: POST /admin/users
    - Request: GET /admin/users
    - Request: PATCH /admin/profile

### 2. Set Environment Variables
```javascript
{
  "base_url": "http://localhost:8080",
  "token": "",
  "user_id": "",
  "test_username": "test_user_" + new Date().getTime()
}
```

### 3. Pre-request Scripts
```javascript
// Auto-append timestamp to avoid duplicate usernames
pm.environment.set("test_username", "user_" + new Date().getTime());
```

### 4. Test Scripts (pada Tests tab)
```javascript
// Save token dari login response
if (pm.response.code === 200 && pm.response.json().token) {
    pm.environment.set("token", pm.response.json().token);
    pm.environment.set("user_id", pm.response.json().user.id);
}

// Verify response status
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});

// Verify token format
pm.test("Token is JWT format", function () {
    var token = pm.response.json().token;
    pm.expect(token).to.match(/^eyJ/); // JWT starts with eyJ
});
```

---

## 🧪 Unit Test Template (untuk future)

```go
package handler

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestAdminSetup(t *testing.T) {
    // Arrange
    req := model.AdminSetupRequest{
        Username: "testadmin",
        Password: "password123",
        FullName: "Test Admin",
    }
    
    // Act
    resp, err := h.AuthUsecase.AdminSetup(ctx, &req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Equal(t, "super_admin", resp.User.Role)
}
```

---

## ✅ Sign-Off Checklist

Setelah semua test passed:

- [ ] All 14 main test cases passed
- [ ] All error scenarios handled
- [ ] Response format consistent
- [ ] Swagger documentation updated
- [ ] POSTMAN_EXAMPLES.md verified
- [ ] Token auto-save working in Postman
- [ ] Role-based access control working
- [ ] Database queries optimized
- [ ] Code compiled without errors
- [ ] Docker containers healthy

---

## 📞 Support Resources

- **Swagger UI:** `http://localhost:8080/swagger/index.html`
- **Backend Logs:** `docker logs jbm_api`
- **Database Logs:** `docker logs jbm_postgres`
- **Redis Logs:** `docker logs jbm_redis`

---

**Last Updated:** 2026-04-18  
**Status:** Ready for Integration Testing  
**Next Phase:** Order & Trip Endpoints Testing
