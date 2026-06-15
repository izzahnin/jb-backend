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

// RegisterOrderRoutes handles commercial order lifecycle used by Admin Sales.
func (h *Handler) RegisterOrderRoutes(r *gin.RouterGroup) {
	r.POST("/admin/orders", h.CreateOrder)
	r.GET("/admin/orders", h.ListOrders)
	r.GET("/admin/orders/:id", h.GetOrder)
	r.PATCH("/admin/orders/:id", h.UpdateOrderStatus)
	r.DELETE("/admin/orders/:id", h.CancelOrder)
}

// CreateOrder membuat order baru untuk customer (admin_sales only).
// @Summary Create new commercial order
// @Description Create a new order with customer, origin, destination, and total containers. Status auto-set to 'pending'. Admin ID extracted from JWT.
// @Tags Orders
// @Accept json
// @Produce json
// @Param body body model.CreateOrderRequest true "Order details: customer_id (required), origin (required), destination (required), total_containers (required)"
// @Success 201 {object} map[string]interface{} "Order created successfully with id, order_number, customer_id, admin_id, origin, destination, total_containers, order_date, status"
// @Failure 400 {object} map[string]string "Bad request: invalid JSON or missing required fields"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_sales role required"
// @Failure 404 {object} map[string]string "Not found - customer_id tidak ada"
// @Failure 422 {object} map[string]string "Unprocessable entity - failed to create order"
// @Router /admin/orders [post]
// @Security Bearer
func (h *Handler) CreateOrder(c *gin.Context) {
	var input model.CreateOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user in context"})
		return
	}
	adminID, ok := userIDVal.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user in context"})
		return
	}

	// Create Order struct for database operations (will set auto-generated fields)
	order := &model.Order{
		CustomerID:      input.CustomerID,
		Origin:          input.Origin,
		Destination:     input.Destination,
		TotalContainers: input.TotalContainers,
		AdminID:         adminID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.OrderUsecase.CreateOrder(ctx, order); err != nil {
		switch err {
		case usecase.ErrOrderOriginRequired, usecase.ErrOrderDestinationRequired,
			usecase.ErrOrderCustomerRequired, usecase.ErrOrderTotalContainersInvalid:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order berhasil dibuat", "data": order})
}

// ListOrders mengambil daftar semua orders dengan pagination (admin_sales only).
// @Summary List all orders
// @Description Get all orders with pagination support. Returns list of orders sorted by order_date DESC.
// @Tags Orders
// @Accept json
// @Produce json
// @Param limit query int false "Items per page (default 10)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {object} map[string]interface{} "Order list with count and pagination info"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_sales role required"
// @Failure 500 {object} map[string]string "Internal server error"
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

// GetOrder mengambil detail satu order berdasarkan ID (admin_sales only).
// @Summary Get order details
// @Description Retrieve detailed information of a specific order by ID. Includes customer, admin, and status info.
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{} "Order details with all fields"
// @Failure 400 {object} map[string]string "Bad request - invalid ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_sales role required"
// @Failure 404 {object} map[string]string "Not found - order_id tidak ada"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/orders/{id} [get]
// @Security Bearer
func (h *Handler) GetOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
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

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// UpdateOrderStatus mengubah status order (admin_sales only).
// @Summary Update order status
// @Description Update the status of an order. Valid transitions: pending → partial/completed/cancelled, partial → completed/cancelled.
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param body body model.UpdateOrderRequest true "Status update: status (required, one of: pending, partial, completed, cancelled)"
// @Success 200 {object} map[string]interface{} "Order status updated successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid ID or status"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_sales role required"
// @Failure 404 {object} map[string]string "Not found - order_id tidak ada"
// @Failure 409 {object} map[string]string "Conflict - invalid status transition"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/orders/{id} [patch]
// @Security Bearer
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrOrderInvalidID.Error()})
		return
	}

	var input model.UpdateOrderRequest
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

// CancelOrder membatalkan order (admin_sales only).
// @Summary Cancel order
// @Description Cancel a specific order. Order must be in 'pending' or 'partial' status to be cancelled.
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]string "Order cancelled successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_sales role required"
// @Failure 404 {object} map[string]string "Not found - order_id tidak ada"
// @Failure 409 {object} map[string]string "Conflict - cannot cancel order (invalid status or already completed)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/orders/{id} [delete]
// @Security Bearer
func (h *Handler) CancelOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
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
