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
	r.GET("/admin/trips", h.ListTrips)
	r.POST("/admin/trips", h.CreateTrip)
	r.GET("/admin/trips/:id", h.GetTrip)
	r.PATCH("/admin/trips/:id/start", h.StartTrip)
	r.PATCH("/admin/trips/:id/deliver", h.CompleteTrip)
}

// ListTrips mengambil daftar semua trips (admin_ops only).
// @Summary List all trips
// @Description Retrieve all trips (surat jalan) with their status, truck, and driver assignments. Sorted by creation date (newest first).
// @Tags Trips
// @Produce json
// @Success 200 {object} map[string]interface{} "List of all trips with count"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_ops role required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/trips [get]
// @Security Bearer
func (h *Handler) ListTrips(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	trips, err := h.TripUsecase.GetAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Trips retrieved successfully",
		"count":   len(trips),
		"data":    trips,
	})
}

// CreateTrip membuat trip baru untuk order (admin_ops only).
// @Summary Create new trip
// @Description Create a new trip (surat jalan) for an order. Requires order_id, truck_id, and driver_id. Trip number is generated automatically. Allocates truck and driver for the order.
// @Tags Trips
// @Accept json
// @Produce json
// @Param body body model.CreateTripRequest true "Trip details: order_id (required), truck_id (required), driver_id (required)"
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
		OrderID:  input.OrderID,
		TruckID:  input.TruckID,
		DriverID: input.DriverID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.TripUsecase.CreateTrip(ctx, trip, actorUserID); err != nil {
		switch err {
		case usecase.ErrOrderInvalidID, usecase.ErrTruckInvalidID, usecase.ErrDriverInvalidID:
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

// GetTrip mengambil detail trip berdasarkan ID (admin_ops only).
// @Summary Get trip by ID
// @Description Retrieve a single trip (surat jalan) by trip ID. Shows trip number, status, truck, and driver assignments.
// @Tags Trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} map[string]interface{} "Trip detail"
// @Failure 400 {object} map[string]string "Bad request - invalid trip ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - admin_ops role required"
// @Failure 404 {object} map[string]string "Not found - trip tidak ditemukan"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/trips/{id} [get]
// @Security Bearer
func (h *Handler) GetTrip(c *gin.Context) {
	tripID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrTripInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	trip, err := h.TripUsecase.GetByID(ctx, tripID)
	if err != nil {
		if err == usecase.ErrTripInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if trip == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrTripNotFound.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": trip})
}

// StartTrip memulai trip dan mengubah status ke in_transit (admin_ops only).
// @Summary Start trip
// @Description Start a trip by setting container_number and seal_number. Changes trip status from 'assigned' to 'in_transit'. Real-time tracking begins.
// @Tags Trips
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param body body model.StartTripRequest true "Trip start details: container_number and seal_number (both required)"
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

	var input model.StartTripRequest
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
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
