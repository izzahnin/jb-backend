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

	// Ambil info armada jika ada
	var truckInfo interface{} = nil
	if order.TruckID != nil {
		truck, err := h.TruckUsecase.GetByID(ctx, *order.TruckID)
		if err == nil {
			truckInfo = gin.H{
				"id":           truck.ID,
				"plate_number": truck.PlateNumber,
				"driver_name":  truck.DriverName,
			}
		}
	}

	// Ambil lokasi terbaru armada jika ada
	var locationInfo interface{} = nil
	if order.TruckID != nil {
		location, err := h.LocRepo.GetLatestLocation(ctx, *order.TruckID)
		if err == nil {
			locationInfo = location
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"order": gin.H{
			"order_number": order.OrderNumber,
			"status":       order.Status,
			"origin":       order.Origin,
			"destination":  order.Destination,
			"created_at":   order.CreatedAt,
		},
		"truck":    truckInfo,
		"location": locationInfo,
	})
}
