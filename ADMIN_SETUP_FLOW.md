# Admin Setup & User Management Flow

> **📌 Note**: In Swagger UI, `/admin/setup` endpoint is now in **"User Management"** group (not Authentication). Check http://localhost:8080/swagger/docs - it's organized by feature groups for easier navigation.

## Total 3 Role Types

| Role | Permissions | Can Create Users | Use Case |
|------|------------|-----------------|----------|
| **super_admin** | • Manage all users<br>• Create orders<br>• Manage trips & trucks | ✅ YES | System administrator |
| **admin_sales** | • Create & view orders | ❌ NO | Sales team member |
| **admin_ops** | • Manage trucks<br>• Create & manage trips<br>• Track GPS locations | ❌ NO | Operations team member |

---

## Complete Setup Workflow

### Phase 1: First-Time Setup (One-Time Only)
**Endpoint:** `POST /admin/setup` (PUBLIC - no authentication needed)

Create the first super_admin account. **This endpoint works only ONCE** - after the first admin is created, attempting to call it again will return error.

```bash
curl -X POST http://localhost:8080/admin/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "superadmin",
    "password": "password123",
    "full_name": "System Administrator"
  }'
```

**Response (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1718726400,
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

**⚠️ Important:** Save the JWT token! Use it for the next requests.

---

### Phase 2: Super Admin Login (After Restart)
**Endpoint:** `POST /auth/login` (PUBLIC)

If the super_admin needs to login again (e.g., after server restart), use regular login endpoint:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "superadmin",
    "password": "password123"
  }'
```

**Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": 1718812800,
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

---

### Phase 3: Super Admin Creates Other Admin Accounts
**Endpoint:** `POST /admin/users` (PROTECTED - super_admin only)

The super_admin can now create accounts for admin_sales or admin_ops team members.

#### Example 1: Create admin_sales user

```bash
curl -X POST http://localhost:8080/admin/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "username": "sales_user",
    "full_name": "John Sales",
    "password": "sales123456",
    "role": "admin_sales",
    "is_active": true
  }'
```

**Response (201 Created):**
```json
{
  "message": "User berhasil dibuat",
  "data": {
    "id": 2,
    "username": "sales_user",
    "full_name": "John Sales",
    "role": "admin_sales",
    "is_active": true,
    "created_at": "2026-04-18T10:05:00Z"
  }
}
```

#### Example 2: Create admin_ops user

```bash
curl -X POST http://localhost:8080/admin/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "username": "ops_user",
    "full_name": "Jane Operations",
    "password": "ops123456",
    "role": "admin_ops",
    "is_active": true
  }'
```

**Response (201 Created):**
```json
{
  "message": "User berhasil dibuat",
  "data": {
    "id": 3,
    "username": "ops_user",
    "full_name": "Jane Operations",
    "role": "admin_ops",
    "is_active": true,
    "created_at": "2026-04-18T10:10:00Z"
  }
}
```

#### Example 3: Create another super_admin (if needed)

```bash
curl -X POST http://localhost:8080/admin/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "username": "superadmin2",
    "full_name": "Secondary Administrator",
    "password": "admin123456",
    "role": "super_admin",
    "is_active": true
  }'
```

---

### Phase 4: View All Users
**Endpoint:** `GET /admin/users` (PROTECTED - super_admin only)

List all admin accounts in the system.

```bash
curl -X GET http://localhost:8080/admin/users \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response (200 OK):**
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
      "username": "sales_user",
      "full_name": "John Sales",
      "role": "admin_sales",
      "is_active": true,
      "created_at": "2026-04-18T10:05:00Z"
    },
    {
      "id": 3,
      "username": "ops_user",
      "full_name": "Jane Operations",
      "role": "admin_ops",
      "is_active": true,
      "created_at": "2026-04-18T10:10:00Z"
    }
  ]
}
```

---

## JWT Token Structure

The JWT token contains claims with user information for authorization checks:

```json
{
  "user_id": 1,
  "username": "superadmin",
  "role": "super_admin",
  "exp": 1718812800,
  "iat": 1718726400
}
```

**Token Lifespan:** 24 hours from login/setup

---

## Access Control Rules

### Which endpoint can super_admin access?

✅ super_admin can access:
- `GET /admin/users` (list all users)
- `POST /admin/users` (create new users)
- `POST /admin/orders` (create orders - sales permission)
- `GET /admin/orders` (view orders - sales permission)
- `POST /admin/trips` (create trips - ops permission)
- `GET /admin/trips` (list trips - ops permission)
- `GET /admin/orders/:id/trips` (view trips - ops permission)
- `PATCH /admin/trips/:id/start` (start trip - ops permission)
- `PATCH /admin/trips/:id/deliver` (complete trip - ops permission)
- `GET /admin/trucks` (view trucks - ops permission)
- `POST /admin/trucks` (create truck - ops permission)
- `DELETE /admin/trucks/:id` (deactivate truck - ops permission)
- `GET /admin/dashboard/stats` (view dashboard - all admin can access)

### Which endpoint can admin_sales access?

✅ admin_sales can access:
- `POST /admin/orders` (create orders)
- `GET /admin/orders` (view orders)
- `GET /admin/orders/:id` (view order details)
- `PATCH /admin/orders/:id` (update order status)
- `GET /admin/dashboard/stats` (view dashboard)

❌ admin_sales CANNOT:
- Create/manage users
- Create/manage trips
- Create/manage trucks

### Which endpoint can admin_ops access?

✅ admin_ops can access:
- `GET /admin/trucks` (view trucks)
- `POST /admin/trucks` (create trucks)
- `PATCH /admin/trucks/:id` (update truck)
- `DELETE /admin/trucks/:id` (deactivate truck)
- `POST /admin/trips` (create trips)
- `GET /admin/trips` (list trips)
- `GET /admin/orders/:id/trips` (view trips for order)
- `PATCH /admin/trips/:id/start` (start trip - dispatch)
- `PATCH /admin/trips/:id/deliver` (complete trip)
- `POST /trips/:id/location` (save GPS location)
- `GET /trips/:id/location` (get latest location)
- `GET /trips/:id/locations` (get GPS history)
- `GET /admin/dashboard/stats` (view dashboard)

❌ admin_ops CANNOT:
- Create/manage users
- Create orders

---

## Troubleshooting

### "Admin user already exists" error on POST /admin/setup
**Cause:** The first super_admin has already been created.
**Solution:** Use `POST /auth/login` to login with existing super_admin credentials.

### "Unauthorized" error on POST /admin/users
**Cause:** Missing or invalid JWT token in Authorization header.
**Solution:** 
1. Ensure `Authorization: Bearer <token>` header is present
2. Verify token is not expired (24-hour lifespan)
3. Login again with `POST /auth/login` to get fresh token

### "Forbidden - super_admin role required" error on POST /admin/users
**Cause:** Only super_admin can create users. The logged-in user is admin_sales or admin_ops.
**Solution:** Use a super_admin account to create other users.

### "role must be one of: super_admin, admin_sales, admin_ops" error
**Cause:** Invalid role value in CreateUserRequest.
**Solution:** Use only one of: `super_admin`, `admin_sales`, `admin_ops`

---

## Username vs Full Name (Best Practice)

### Real World Practice:
- **Username** = Technical identifier (like "john.sales" or "jane.ops")
  - Used for login
  - Unique & immutable (never changes)
  - Technical/programmatic use
  - Visible in audit logs & backend systems

- **Full Name** = Display name (like "John Sales" or "Jane Operations")
  - Used for UI/reports
  - Can be changed anytime
  - User-friendly
  - Shows in dashboards, notifications, audit trails

### Workflow:
1. **Setup / Create User**: Admin provides both username (for login) and optional full_name (for display)
2. **If full_name not provided**: System defaults to username (so display still works)
3. **Update Profile**: Users can change their full_name later via `PATCH /admin/profile`

---

## Update Own Profile

Users can update their own profile information (fullname and/or password) anytime.

**Endpoint:** `PATCH /admin/profile` (PROTECTED - all authenticated admins)

```bash
curl -X PATCH http://localhost:8080/admin/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "full_name": "John Smith",
    "password": "newpassword123456"
  }'
```

**Note:** Can update either field or both. At least one must be provided.

**Response (200 OK):**
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

### Example 1: Update only full_name
```bash
curl -X PATCH http://localhost:8080/admin/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "full_name": "Bambang Suryanto"
  }'
```

### Example 2: Update only password
```bash
curl -X PATCH http://localhost:8080/admin/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "password": "mynewpassword123456"
  }'
```

### Example 3: Update both
```bash
curl -X PATCH http://localhost:8080/admin/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "full_name": "Bambang Suryanto",
    "password": "mynewpassword123456"
  }'
```

---

## Setup Options

### Option 1: Setup dengan full_name (Recommended)
```bash
curl -X POST http://localhost:8080/admin/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "superadmin",
    "password": "password123",
    "full_name": "System Administrator"
  }'
```

### Option 2: Setup tanpa full_name (full_name akan = username)
```bash
curl -X POST http://localhost:8080/admin/setup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "superadmin",
    "password": "password123"
  }'
```

---

## Key Takeaways

1. **First-time setup is one-time only** - creates super_admin via `POST /admin/setup`
2. **Super admin then creates other admins** - via `POST /admin/users` with appropriate role
3. **Username** = technical identifier (never changes), **Full Name** = display name (can change)
4. **Full Name defaults to Username** if not provided during setup/creation
5. **Users can update their own profile** via `PATCH /admin/profile` (fullname and/or password)
6. **All admin operations require JWT token** in `Authorization: Bearer` header
7. **Role hierarchy is enforced by middleware** - endpoint returns 403 Forbidden if role doesn't match
8. **Tokens expire after 24 hours** - must re-login with `POST /auth/login` to get fresh token
