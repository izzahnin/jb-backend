package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes mendaftarkan 1 endpoint public (no auth required).
// - GET /public/orders/:order_number/track - Track order oleh customer (public/anonymous)
func (h *Handler) RegisterPublicRoutes(r *gin.Engine) {
	r.GET("/public/orders/:order_number/track", h.PublicOrderTracking)
}

// PublicOrderTracking mengambil informasi tracking order untuk customer.
// @Summary Track public order
// @Description Track order status and assignment without authentication. Returns order status, truck info, and real-time location
// @Tags Public
// @Accept json
// @Produce json
// @Param order_number path string true "Order number (e.g., ORD-001)"
// @Success 200 {object} object "Order tracking data with truck and location info"
// @Failure 400 {object} map[string]string "Invalid order number"
// @Failure 404 {object} map[string]string "Order not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /public/orders/{order_number}/track [get]
func (h *Handler) PublicOrderTracking(c *gin.Context) {
	orderNumber := c.Param("order_number")
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor order tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	order, err := h.OrderUsecase.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order tidak ditemukan"})
		return
	}

	trips, err := h.TripUsecase.GetByOrderID(ctx, order.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data trip"})
		return
	}

	tripTracking := make([]gin.H, 0, len(trips))
	for _, trip := range trips {
		latestLoc, _ := h.LocationUsecase.GetLatest(ctx, trip.ID)
		tripTracking = append(tripTracking, gin.H{
			"trip":            trip,
			"latest_location": latestLoc,
		})
	}

	// Return order data in the expected frontend format
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"order": order,
			"trips": tripTracking,
		},
	})
}
