package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/usecase"
)

// RegisterTripRoutes handles operational execution (surat jalan) managed by Admin Ops.
func (h *Handler) RegisterTripRoutes(r *gin.RouterGroup) {
	r.POST("/admin/trips", h.CreateTrip)
	r.GET("/admin/orders/:id/trips", h.ListTripsByOrder)
	r.PATCH("/admin/trips/:id/start", h.StartTrip)
	r.PATCH("/admin/trips/:id/deliver", h.CompleteTrip)
}

// CreateTrip membuat trip baru untuk order (admin_ops only).
// @Summary Create new trip
// @Description Create a new trip (surat jalan) for an order. Requires order_id, truck_id, driver_id, and trip_number. Allocates truck and driver for the order.
// @Tags Trips
// @Accept json
// @Produce json
// @Param body body model.CreateTripRequest true "Trip details: order_id (required), truck_id (required), driver_id (required), trip_number (required)"
// @Success 201 {object} map[string]interface{} "Trip created successfully with trip_number, status (assigned), truck and driver allocated"
// @Failure 400 {object} map[string]string "Bad request - invalid JSON or missing required fields"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_ops role required"
// @Failure 404 {object} map[string]string "Not found - order_id, truck_id, or driver_id tidak ada"
// @Failure 409 {object} map[string]string "Conflict - truck/driver inactive or invalid order status"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/trips [post]
// @Security Bearer
func (h *Handler) CreateTrip(c *gin.Context) {
	var input model.CreateTripRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user in context"})
		return
	}
	actorUserID, ok := userIDVal.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user in context"})
		return
	}

	// Create Trip struct for database operations (will set auto-generated fields)
	trip := &model.Trip{
		OrderID:    input.OrderID,
		TruckID:    input.TruckID,
		DriverID:   input.DriverID,
		TripNumber: input.TripNumber,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.TripUsecase.CreateTrip(ctx, trip, actorUserID); err != nil {
		switch err {
		case usecase.ErrOrderInvalidID, usecase.ErrTruckInvalidID, usecase.ErrDriverInvalidID, usecase.ErrTripNumberRequired:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrOrderNotFound, usecase.ErrTruckNotFound, usecase.ErrDriverNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case usecase.ErrTruckInactive, usecase.ErrDriverInactive, usecase.ErrOrderInvalidStatusTransition:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Trip berhasil dibuat", "data": trip})
}

// ListTripsByOrder mengambil daftar trips untuk satu order (admin_ops only).
// @Summary List trips for order
// @Description Retrieve all trips (surat jalan) associated with a specific order ID. Shows trip number, status, truck, and driver assignments.
// @Tags Trips
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]interface{} "List of trips for the order"
// @Failure 400 {object} map[string]string "Bad request - invalid order ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_ops role required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/orders/{id}/trips [get]
// @Security Bearer
func (h *Handler) ListTripsByOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrOrderInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	trips, err := h.TripUsecase.GetByOrderID(ctx, orderID)
	if err != nil {
		if err == usecase.ErrOrderInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": trips})
}

// StartTrip memulai trip dan mengubah status ke in_transit (admin_ops only).
// @Summary Start trip
// @Description Start a trip by setting container_number and seal_number. Changes trip status from 'assigned' to 'in_transit'. Real-time tracking begins.
// @Tags Trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param body body map[string]string true "Trip start details: container_number (required), seal_number (required)"
// @Success 200 {object} map[string]string "Trip started, status changed to in_transit"
// @Failure 400 {object} map[string]string "Bad request - invalid ID or missing container/seal numbers"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_ops role required"
// @Failure 404 {object} map[string]string "Not found - trip_id tidak ada"
// @Failure 409 {object} map[string]string "Conflict - cannot start trip (invalid status or already in_transit)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/trips/{id}/start [patch]
// @Security Bearer
func (h *Handler) StartTrip(c *gin.Context) {
	tripID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrTripInvalidID.Error()})
		return
	}

	var input struct {
		ContainerNumber string `json:"container_number"`
		SealNumber      string `json:"seal_number"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user in context"})
		return
	}
	actorUserID := userIDVal.(int)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.TripUsecase.StartTrip(ctx, tripID, input.ContainerNumber, input.SealNumber, actorUserID); err != nil {
		switch err {
		case usecase.ErrTripInvalidID, usecase.ErrTripContainerRequired, usecase.ErrTripSealRequired:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrTripNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case usecase.ErrTripInvalidStatusTransition:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trip berangkat dan status menjadi in_transit"})
}

// CompleteTrip menyelesaikan trip dan mengubah status ke delivered (admin_ops only).
// @Summary Complete trip
// @Description Complete a trip by marking it as delivered. Changes trip status from 'in_transit' to 'delivered'. Order status automatically updated to 'completed'.
// @Tags Trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} map[string]string "Trip completed, status changed to delivered, order marked as completed"
// @Failure 400 {object} map[string]string "Bad request - invalid trip ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_ops role required"
// @Failure 404 {object} map[string]string "Not found - trip_id tidak ada"
// @Failure 409 {object} map[string]string "Conflict - cannot complete trip (invalid status or not in_transit)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/trips/{id}/deliver [patch]
// @Security Bearer
func (h *Handler) CompleteTrip(c *gin.Context) {
	tripID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrTripInvalidID.Error()})
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user in context"})
		return
	}
	actorUserID := userIDVal.(int)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.TripUsecase.CompleteTrip(ctx, tripID, actorUserID); err != nil {
		switch err {
		case usecase.ErrTripInvalidID:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrTripNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case usecase.ErrTripInvalidStatusTransition:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trip selesai dan status menjadi delivered"})
}
