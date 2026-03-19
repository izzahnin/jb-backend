package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/usecase"
)

// RegisterLocationRoutes mendaftarkan 3 endpoint tracking lokasi armada (public, no auth).
// - POST /trucks/:id/location - Catat lokasi GPS armada
// - GET /trucks/:id/location - Dapatkan lokasi terbaru armada
// - GET /trucks/:id/locations - Dapatkan riwayat lokasi armada (limit=50)
func (h *Handler) RegisterLocationRoutes(r *gin.Engine) {
	r.POST("/trucks/:id/location", h.PostLocation)
	r.GET("/trucks/:id/locations", h.GetLocationHistory)
	r.GET("/trucks/:id/location", h.GetLatestLocation)
}

// PostLocation mencatat posisi GPS armada ke Redis (geo-spatial data).
// @Summary Record truck location
// @Description Record GPS coordinates and timestamp for a specific truck for real-time tracking
// @Tags Locations
// @Accept json
// @Produce json
// @Param id path int true "Truck ID"
// @Param body body object true "Location data" example({"lat":-8.5,"lon":120.7,"ts":"2026-03-16T11:45:00Z"})
// @Success 200 {object} map[string]string "Location saved successfully"
// @Failure 400 {object} map[string]string "Invalid truck ID or JSON format"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /trucks/{id}/location [post]
func (h *Handler) PostLocation(c *gin.Context) {
	idStr := c.Param("id")
	truckID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID truck tidak valid"})
		return
	}

	var input struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
		Ts  string  `json:"ts"`
	}
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

	if err := h.LocationUsecase.SaveLocation(ctx, truckID, input.Lat, input.Lon, ts); err != nil {
		if err == usecase.ErrLocationInvalidTruckID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lokasi berhasil disimpan"})
}

// GetLocationHistory mengambil riwayat lokasi armada (terbatas default 50 entries).
// @Summary Get truck location history
// @Description Retrieve location history for a specific truck with pagination support
// @Tags Locations
// @Accept json
// @Produce json
// @Param id path int true "Truck ID"
// @Param limit query int false "Number of records (max 200)" default(50)
// @Success 200 {object} object "Location history data"
// @Failure 400 {object} map[string]string "Invalid truck ID"
// @Failure 404 {object} map[string]string "Truck not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /trucks/{id}/locations [get]
func (h *Handler) GetLocationHistory(c *gin.Context) {
	idStr := c.Param("id")
	truckID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID truck tidak valid"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 200 {
		limit = 50
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = h.TruckUsecase.GetByID(ctx, truckID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Truck tidak ditemukan"})
		return
	}

	locations, err := h.LocRepo.GetHistory(ctx, truckID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil history lokasi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": locations})
}

// GetLatestLocation mengambil lokasi terbaru armada (real-time position).
// @Summary Get latest truck location
// @Description Retrieve the most recent location data for a specific truck in real-time
// @Tags Locations
// @Accept json
// @Produce json
// @Param id path int true "Truck ID"
// @Success 200 {object} object "Latest location data"
// @Failure 400 {object} map[string]string "Invalid truck ID"
// @Failure 404 {object} map[string]string "Truck or location not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /trucks/{id}/location [get]
func (h *Handler) GetLatestLocation(c *gin.Context) {
	idStr := c.Param("id")
	truckID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID truck tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = h.TruckUsecase.GetByID(ctx, truckID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Truck tidak ditemukan"})
		return
	}

	location, err := h.LocRepo.GetLatestLocation(ctx, truckID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lokasi truck tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": location})
}
