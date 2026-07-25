package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterDashboardRoutes mendaftarkan admin dashboard endpoints (all authenticated admins).
// GET /admin/dashboard/stats - Get dashboard statistics (defined here)
// PATCH /admin/profile        - Update own profile (defined in user_routes.go, registered here for convenience)
func (h *Handler) RegisterDashboardRoutes(r *gin.RouterGroup) {
	r.GET("/admin/dashboard/stats", h.GetDashboardStats)
}

// DashboardStats adalah response model untuk dashboard stats.
type DashboardStats struct {
	TotalOrders    int `json:"total_orders"`
	TotalTrucks    int `json:"total_trucks"`
	TotalUsers     int `json:"total_users"`
	TotalAdmins    int `json:"total_admins"`
	ActiveTrucks   int `json:"active_trucks"`
	OrderBreakdown struct {
		Pending   int `json:"pending"`
		Partial   int `json:"partial"`
		Completed int `json:"completed"`
		Cancelled int `json:"cancelled"`
	} `json:"order_breakdown"`
}

// GetDashboardStats mengambil statistik dashboard untuk admin.
// @Summary Get dashboard statistics
// @Description Get dashboard overview with order stats, truck count, user count, etc.
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} DashboardStats "Dashboard statistics"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/dashboard/stats [get]
// @Security Bearer
func (h *Handler) GetDashboardStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	stats := &DashboardStats{}

	// Active Orders (pending + partial only)
	totalOrders, err := h.OrderRepo.CountActiveOrders(ctx)
	if err == nil {
		stats.TotalOrders = totalOrders
	}

	// Total Trucks
	totalTrucks, err := h.TruckRepo.Count(ctx)
	if err == nil {
		stats.TotalTrucks = totalTrucks
	}

	// Active Trucks
	activeTrucks, err := h.TruckRepo.CountActive(ctx)
	if err == nil {
		stats.ActiveTrucks = activeTrucks
	}

	// Total Users
	totalUsers, err := h.UserRepo.Count(ctx)
	if err == nil {
		stats.TotalUsers = totalUsers
	}

	// Total Admins
	totalAdmins, err := h.UserRepo.CountByRole(ctx, "super_admin")
	if err == nil {
		stats.TotalAdmins = totalAdmins
	}

	// Order Breakdown by Status
	pending, _ := h.OrderRepo.CountByStatus(ctx, "pending")
	partial, _ := h.OrderRepo.CountByStatus(ctx, "partial")
	completed, _ := h.OrderRepo.CountByStatus(ctx, "completed")
	cancelled, _ := h.OrderRepo.CountByStatus(ctx, "cancelled")

	stats.OrderBreakdown.Pending = pending
	stats.OrderBreakdown.Partial = partial
	stats.OrderBreakdown.Completed = completed
	stats.OrderBreakdown.Cancelled = cancelled

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
