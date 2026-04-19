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

// RegisterLocationRoutes registers trip-based tracking endpoints.
func (h *Handler) RegisterLocationRoutes(r *gin.RouterGroup) {
	r.POST("/trips/:id/location", h.PostLocation)
	r.GET("/trips/:id/location", h.GetLatestLocation)
	r.GET("/trips/:id/locations", h.GetLocationHistory)
}

// PostLocation menyimpan GPS coordinate untuk trip real-time tracking.
// @Summary Save trip location
// @Description Save current GPS location for a trip. Accepts latitude, longitude, optional speed, and timestamp. Used for real-time vehicle tracking.
// @Tags Locations
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param body body model.CreateLocationRequest true "Location data: lat (float64 required), lon (float64 required), speed (float64 optional), ts (RFC3339 string optional, defaults to now)"
// @Success 200 {object} map[string]string "Location saved successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid trip ID or invalid lat/lon coordinates"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /trips/{id}/location [post]
func (h *Handler) PostLocation(c *gin.Context) {
	tripID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrLocationInvalidTripID.Error()})
		return
	}

	var input model.CreateLocationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ts := time.Now().UTC()
	if input.Ts != "" {
		if parsed, err := time.Parse(time.RFC3339, input.Ts); err == nil {
			ts = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.LocationUsecase.SaveLocation(ctx, tripID, input.Lat, input.Lon, input.Speed, ts); err != nil {
		switch err {
		case usecase.ErrLocationInvalidTripID, usecase.ErrLocationInvalidCoords:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lokasi trip berhasil disimpan"})
}

// GetLocationHistory mengambil history lokasi dengan limit (default 50, max 500).
// @Summary Get location history
// @Description Retrieve the location history for a trip. Returns all saved GPS coordinates in descending order (latest first). Supports pagination via limit parameter.
// @Tags Locations
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Param limit query int false "Max number of locations to return (default 50, max 500)"
// @Success 200 {object} map[string]interface{} "Array of location records with lat, lon, speed, timestamp"
// @Failure 400 {object} map[string]string "Bad request - invalid trip ID format"
// @Failure 500 {object} map[string]string "Internal server error - failed to retrieve location history"
// @Router /trips/{id}/locations [get]
func (h *Handler) GetLocationHistory(c *gin.Context) {
	tripID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrLocationInvalidTripID.Error()})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	locations, err := h.LocationUsecase.GetHistory(ctx, tripID, limit)
	if err != nil {
		if err == usecase.ErrLocationInvalidTripID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil history lokasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": locations})
}

// GetLatestLocation mengambil lokasi terbaru untuk trip.
// @Summary Get latest location
// @Description Retrieve the most recent GPS location for a trip. Returns the latest coordinate point recorded.
// @Tags Locations
// @Accept json
// @Produce json
// @Param id path int true "Trip ID"
// @Success 200 {object} map[string]interface{} "Latest location record with lat, lon, speed, timestamp"
// @Failure 400 {object} map[string]string "Bad request - invalid trip ID format"
// @Failure 404 {object} map[string]string "Not found - no location recorded for this trip"
// @Failure 500 {object} map[string]string "Internal server error - failed to retrieve latest location"
// @Router /trips/{id}/location [get]
func (h *Handler) GetLatestLocation(c *gin.Context) {
	tripID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrLocationInvalidTripID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	location, err := h.LocationUsecase.GetLatest(ctx, tripID)
	if err != nil {
		if err == usecase.ErrLocationInvalidTripID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil lokasi terbaru"})
		return
	}
	if location == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lokasi trip tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": location})
}
