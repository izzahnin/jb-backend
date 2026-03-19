package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/usecase"
	"github.com/izzahnin/jalur-berlian-backend/pkg/helper"
)

// RegisterOrderRoutes mendaftarkan 8 endpoint manajemen order.
// - POST /admin/orders - Buat order baru (status=pending)
// - GET /admin/orders - List order (dengan pagination)
// - GET /admin/orders/:id - Detail order
// - PATCH /admin/orders/assign - Assign armada ke order
// - POST /admin/orders/:id/confirm-pickup - Konfirmasi pickup (pickup→in_transit)
// - POST /admin/orders/:id/confirm-delivery - Konfirmasi delivery (in_transit→delivered)
// - PATCH /admin/orders/:id - Update status order
// - DELETE /admin/orders/:id - Cancel order
func (h *Handler) RegisterOrderRoutes(r *gin.RouterGroup) {
	r.POST("/admin/orders", h.CreateOrder)
	r.GET("/admin/orders", h.ListOrders)
	r.GET("/admin/orders/:id", h.GetOrder)
	r.PATCH("/admin/orders/assign", h.AssignTruck)
	r.PATCH("/admin/orders/:id", h.UpdateOrderStatus)
	r.DELETE("/admin/orders/:id", h.CancelOrder)
	r.POST("/admin/orders/:id/confirm-pickup", h.ConfirmPickup)
	r.POST("/admin/orders/:id/confirm-delivery", h.ConfirmDelivery)
}

// CreateOrder membuat order baru dengan status "pending".
// @Summary Create a new order
// @Description Create a new order with pending status
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body model.Order true "Order details" example({"order_number":"ORD-202603-001","origin":"Jakarta Pusat","destination":"Bandung Raya"})
// @Success 201 {object} map[string]interface{} "Order created"
// @Failure 400 {object} map[string]string "Invalid input"
// @Router /admin/orders [post]
// @Security Bearer
func (h *Handler) CreateOrder(c *gin.Context) {
	var input model.Order
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.OrderUsecase.CreateOrder(ctx, &input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order berhasil dibuat",
		"data":    input,
	})
}

// ListOrders mengambil list order dengan pagination.
// @Summary List all orders
// @Description Get all orders with pagination support
// @Tags Orders
// @Produce json
// @Param limit query int false "Items per page"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{} "List of orders"
// @Router /admin/orders [get]
// @Security Bearer
func (h *Handler) ListOrders(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pagination := helper.ParsePaginationParams(c)
	offset := pagination.CalculateOffset()

	orders, err := h.OrderRepo.FetchAllWithPagination(ctx, pagination.Limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data order: " + err.Error()})
		return
	}

	totalCount, err := h.OrderRepo.Count(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung total order: " + err.Error()})
		return
	}

	response := pagination.NewPageResponse(orders, totalCount)
	c.JSON(http.StatusOK, response)
}

// GetOrder mengambil detail order berdasarkan ID.
// @Summary Get order details
// @Description Retrieve a specific order by ID
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{} "Order details"
// @Failure 404 {object} map[string]string "Order not found"
// @Router /admin/orders/{id} [get]
// @Security Bearer
func (h *Handler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrOrderInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	order, err := h.OrderUsecase.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// AssignTruck mengassign armada ke order (pending→pickup).
// @Summary Assign truck to order
// @Description Assign a truck to an order
// @Tags Orders
// @Accept json
// @Produce json
// @Param payload body map[string]int true "Order and Truck IDs"
// @Success 200 {object} map[string]string "Truck assigned"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Order or Truck not found"
// @Failure 409 {object} map[string]string "Conflict"
// @Router /admin/orders/assign [patch]
// @Security Bearer
func (h *Handler) AssignTruck(c *gin.Context) {
	var input struct {
		OrderID int `json:"order_id"`
		TruckID int `json:"truck_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.OrderUsecase.AssignTruck(ctx, input.OrderID, input.TruckID); err != nil {
		switch err {
		case usecase.ErrOrderInvalidID, usecase.ErrTruckInvalidID:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrOrderNotFound, usecase.ErrTruckNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case usecase.ErrTruckInactive, usecase.ErrOrderInvalidAssignStatus:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Truk berhasil ditugaskan ke order"})
}

// UpdateOrderStatus mengupdate status order (generic update).
// @Summary Update order status
// @Description Update the status of an order
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param payload body map[string]string true "New status"
// @Success 200 {object} map[string]string "Status updated"
// @Failure 422 {object} map[string]string "Invalid status transition"
// @Router /admin/orders/{id} [patch]
// @Security Bearer
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrOrderInvalidID.Error()})
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.OrderUsecase.UpdateStatus(ctx, id, input.Status); err != nil {
		switch err {
		case usecase.ErrOrderInvalidID:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrOrderNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case usecase.ErrOrderInvalidStatus, usecase.ErrOrderInvalidStatusTransition:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status order berhasil diupdate"})
}

// CancelOrder membatalkan order (any status→cancelled).
// @Summary Cancel order
// @Description Cancel an order
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]string "Order cancelled"
// @Failure 404 {object} map[string]string "Order not found"
// @Failure 409 {object} map[string]string "Cannot cancel order"
// @Router /admin/orders/{id} [delete]
// @Security Bearer
func (h *Handler) CancelOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrOrderInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.OrderUsecase.Cancel(ctx, id); err != nil {
		switch err {
		case usecase.ErrOrderInvalidID:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrOrderNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case usecase.ErrOrderCannotCancel:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order berhasil dibatalkan"})
}

// ConfirmPickup mengkonfirmasi pickup (status: pickup→in_transit).
// Artinya barang sudah diambil dari warehouse dan truck sedang dalam perjalanan.
// @Summary Confirm pickup
// @Description Confirm order pickup (status change: pickup → in_transit)
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{} "Pickup confirmed"
// @Failure 404 {object} map[string]string "Order not found"
// @Failure 409 {object} map[string]string "Invalid status for pickup"
// @Router /admin/orders/{id}/confirm-pickup [post]
// @Security Bearer
func (h *Handler) ConfirmPickup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrOrderInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	order, err := h.OrderUsecase.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if order.Status != "pickup" {
		c.JSON(http.StatusConflict, gin.H{"error": "Order harus dalam status 'pickup' untuk confirm pickup"})
		return
	}

	if err := h.OrderUsecase.UpdateStatus(ctx, id, "in_transit"); err != nil {
		if err == usecase.ErrOrderInvalidStatusTransition {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order berhasil dikonfirmasi pickup (barang sudah diambil)",
		"status":  "in_transit",
	})
}

// ConfirmDelivery mengkonfirmasi delivery (status: in_transit→delivered).
// Artinya barang telah tiba di lokasi tujuan dan diterima customer.
// @Summary Confirm delivery
// @Description Confirm order delivery (status change: in_transit → delivered)
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{} "Delivery confirmed"
// @Failure 404 {object} map[string]string "Order not found"
// @Failure 409 {object} map[string]string "Invalid status for delivery"
// @Router /admin/orders/{id}/confirm-delivery [post]
// @Security Bearer
func (h *Handler) ConfirmDelivery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrOrderInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	order, err := h.OrderUsecase.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if order.Status != "in_transit" {
		c.JSON(http.StatusConflict, gin.H{"error": "Order harus dalam status 'in_transit' untuk confirm delivery"})
		return
	}

	if err := h.OrderUsecase.UpdateStatus(ctx, id, "delivered"); err != nil {
		if err == usecase.ErrOrderInvalidStatusTransition {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order berhasil dikonfirmasi delivery (barang telah diterima)",
		"status":  "delivered",
	})
}
