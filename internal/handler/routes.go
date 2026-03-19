package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/middleware"
)

// RegisterAllRoutes mendaftarkan semua 21 endpoints ke router (main orchestrator).
//
// Organisasi routes:
// ├── Auth Routes (public)      -> auth_handler.go (3 endpoints)
// ├── Admin Routes (with auth)  ->
// │   ├── User Routes           -> user_routes.go (1 endpoint)
// │   ├── Truck Routes          -> truck_routes.go (5 endpoints)
// │   └── Order Routes          -> order_routes.go (8 endpoints)
// ├── Location Routes (public)  -> location_routes.go (3 endpoints)
// └── Public Tracking (public)  -> public_routes.go (1 endpoint)
func (h *Handler) RegisterAllRoutes(r *gin.Engine) {
	// ========================================
	// 3 Auth Endpoints (Public)
	// ========================================
	h.RegisterAuthRoutes(r)

	// ========================================
	// Admin Protected Routes (with JWT + Admin role check)
	// ========================================
	adminRoutes := r.Group("")
	adminRoutes.Use(middleware.AuthMiddleware(h.JWTSecret), middleware.AdminMiddleware())

	// 1 User Management Endpoint
	h.RegisterUserRoutes(adminRoutes)

	// 5 Truck Management Endpoints
	h.RegisterTruckRoutes(adminRoutes)

	// 8 Order Management Endpoints
	h.RegisterOrderRoutes(adminRoutes)

	// ========================================
	// 3 Location Tracking Endpoints (Public, no auth)
	// ========================================
	h.RegisterLocationRoutes(r)

	// ========================================
	// 1 Public Order Tracking Endpoint (Customer view, no auth)
	// ========================================
	h.RegisterPublicRoutes(r)
}
