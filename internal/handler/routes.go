package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/middleware"
)

// RegisterAllRoutes adalah main orchestrator yang mendaftarkan semua endpoints ke router Gin.
//
// Struktur organisasi routes:
// - FOUNDATION LAYERS   : Public & setup routes (no auth required)
// - MIDDLEWARE SETUP    : Admin-protected routes & role-based groups (JWT + role middleware)
// - FEATURE LAYERS      : Feature-based CRUD operations (using middleware groups)
//
// Total ~36 endpoints tersegmentasi berdasarkan:
// 1. Authentication level  : public, protected, role-restricted
// 2. Business capability   : auth, user, customer, driver, truck, order, trip, location
// 3. Access control        : no-auth, admin-all, super-admin, ops-team, sales-team
//
// Route Groups (Total: 11 groups, ~36 endpoints):
//  1. PUBLIC ROUTES (2)           - POST /auth/login, /auth/logout
//  2. ADMIN SETUP (1)             - POST /admin/setup (one-time, public)
//  3. ADMIN DASHBOARD (2)         - GET /admin/dashboard/stats, PATCH /admin/profile
//  4. USER MANAGEMENT (2)         - POST/GET /admin/users (super_admin only)
//  5. CUSTOMER MGMT (5)           - CRUD /admin/customers (all admin roles)
//  6. DRIVER MGMT (5)             - CRUD /admin/drivers (all admin roles)
//  7. TRUCK MGMT (5)              - CRUD /admin/trucks (ops team)
//  8. ORDER MGMT (5)              - CRUD /admin/orders (sales team)
//  9. TRIP MGMT (5)               - CRUD /admin/trips (ops team) + GET /admin/trips/:id
// 10. LOCATION TRACKING (3)       - GET/POST /trips/:id/locations (ops team)
// 11. PUBLIC TRACKING (1)         - GET /order/:order_number/status (public)
func (h *Handler) RegisterAllRoutes(r *gin.Engine) {
	// ===================================================================
	// LAYER 1: PUBLIC ROUTES (No authentication required)
	// ===================================================================
	// Endpoints: POST /auth/login, /auth/logout
	h.RegisterAuthRoutes(r)

	// ===================================================================
	// LAYER 2: PUBLIC SETUP (No authentication, one-time only)
	// ===================================================================
	// Endpoints: POST /admin/setup (create first super_admin account)
	h.RegisterSetupRoutes(r)

	// ===================================================================
	// LAYER 3: ADMIN-PROTECTED ROUTES SETUP (Foundation for all admin routes)
	// ===================================================================
	// Base middleware: JWT authentication + Admin role check
	// All subsequent routes inherit these middlewares
	adminRoutes := r.Group("")
	adminRoutes.Use(middleware.AuthMiddleware(h.JWTSecret), middleware.AdminMiddleware())

	// ===================================================================
	// LAYER 4: ROLE-BASED ROUTE GROUPS (Extends admin protection)
	// ===================================================================
	// Super Admin Group: Full system access (user/admin management)
	// Middleware chain: JWT → AdminRole → SuperAdmin role check
	superAdminRoutes := adminRoutes.Group("")
	superAdminRoutes.Use(middleware.RequireRoles("super_admin"))

	// Operations Group: Truck, trip, location management
	// Middleware chain: JWT → AdminRole → (SuperAdmin OR admin_ops role)
	opsRoutes := adminRoutes.Group("")
	opsRoutes.Use(middleware.RequireRoles("super_admin", "admin_ops"))

	// Sales Group: Order management
	// Middleware chain: JWT → AdminRole → (SuperAdmin OR admin_sales role)
	salesRoutes := adminRoutes.Group("")
	salesRoutes.Use(middleware.RequireRoles("super_admin", "admin_sales"))

	// ===================================================================
	// LAYER 5: ADMIN DASHBOARD (All authenticated admins)
	// ===================================================================
	// Endpoints: GET /admin/dashboard/stats, PATCH /admin/profile
	h.RegisterDashboardRoutes(adminRoutes)
	adminRoutes.PATCH("/admin/profile", h.UpdateProfile)

	// ===================================================================
	// LAYER 6: USER MANAGEMENT (Super admin only)
	// ===================================================================
	// Endpoints: GET/POST /admin/users (list, create admin users)
	h.RegisterUserRoutes(superAdminRoutes)

	// ===================================================================
	// LAYER 7: CUSTOMER MANAGEMENT (All authenticated admins)
	// ===================================================================
	// Endpoints: POST/GET/PATCH/DELETE /admin/customers (5 CRUD operations)
	h.RegisterCustomerRoutes(adminRoutes)

	// ===================================================================
	// LAYER 8: DRIVER MANAGEMENT (All authenticated admins)
	// ===================================================================
	// Endpoints: POST/GET/PATCH/DELETE /admin/drivers (5 CRUD operations)
	h.RegisterDriverRoutes(adminRoutes)

	// ===================================================================
	// LAYER 9: TRUCK MANAGEMENT (Operations team only)
	// ===================================================================
	// Endpoints: POST/GET/PATCH/DELETE /admin/trucks (5 CRUD operations)
	h.RegisterTruckRoutes(opsRoutes)

	// ===================================================================
	// LAYER 10: ORDER MANAGEMENT (Sales team only)
	// ===================================================================
	// Endpoints: POST/GET/PATCH/DELETE /admin/orders (5 CRUD operations)
	h.RegisterOrderRoutes(salesRoutes)

	// ===================================================================
	// LAYER 11: TRIP MANAGEMENT (Operations team only)
	// ===================================================================
	// Endpoints: GET/POST /admin/trips, GET /admin/trips/:id, PATCH /admin/trips/:id/start, /deliver (5 endpoints)
	h.RegisterTripRoutes(opsRoutes)

	// ===================================================================
	// LAYER 12: LOCATION TRACKING (Operations team only)
	// ===================================================================
	// Endpoints: GET/POST /trips/:id/locations, GET /latest (3 endpoints)
	h.RegisterLocationRoutes(opsRoutes)

	// ===================================================================
	// LAYER 13: PUBLIC ORDER TRACKING (No authentication required)
	// ===================================================================
	// Endpoints: GET /order/:order_number/status (public status check)
	h.RegisterPublicRoutes(r)
}
