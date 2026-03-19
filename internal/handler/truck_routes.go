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

// RegisterTruckRoutes mendaftarkan 5 endpoint manajemen armada (truck).
// - POST /admin/trucks - Buat armada baru
// - GET /admin/trucks - List armada (dengan pagination)
// - GET /admin/trucks/:id - Detail armada
// - PUT /admin/trucks/:id - Update armada
// - DELETE /admin/trucks/:id - Deaktifkan armada
func (h *Handler) RegisterTruckRoutes(r *gin.RouterGroup) {
	r.POST("/admin/trucks", h.CreateTruck)
	r.GET("/admin/trucks", h.ListTrucks)
	r.GET("/admin/trucks/:id", h.GetTruck)
	r.PUT("/admin/trucks/:id", h.UpdateTruck)
	r.DELETE("/admin/trucks/:id", h.DeleteTruck)
}

// CreateTruck membuat armada (truck) baru.
// @Summary Create a new truck
// @Description Register a new truck in the fleet
// @Tags Trucks
// @Accept json
// @Produce json
// @Param truck body model.Truck true "Truck details" example({"plate_number":"B-1234-ABC","driver_name":"Budiman Hartono","is_active":true})
// @Success 201 {object} map[string]interface{} "Truck created"
// @Failure 400 {object} map[string]string "Invalid input"
// @Router /admin/trucks [post]
// @Security Bearer
func (h *Handler) CreateTruck(c *gin.Context) {
	var input model.Truck
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.TruckUsecase.RegisterTruck(ctx, &input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Armada berhasil didaftarkan",
		"data":    input,
	})
}

// ListTrucks mengambil list armada dengan pagination.
// @Summary List all trucks
// @Description Get all trucks with pagination support
// @Tags Trucks
// @Produce json
// @Param limit query int false "Items per page (default 10)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} map[string]interface{} "List of trucks"
// @Failure 500 {object} map[string]string "Server error"
// @Router /admin/trucks [get]
// @Security Bearer
func (h *Handler) ListTrucks(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pagination := helper.ParsePaginationParams(c)
	offset := pagination.CalculateOffset()

	trucks, err := h.TruckRepo.FetchAllWithPagination(ctx, pagination.Limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data armada"})
		return
	}

	totalCount, err := h.TruckRepo.Count(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung total armada"})
		return
	}

	response := pagination.NewPageResponse(trucks, totalCount)
	c.JSON(http.StatusOK, response)
}

// GetTruck mengambil detail armada berdasarkan ID.
// @Summary Get truck details
// @Description Retrieve a specific truck by ID
// @Tags Trucks
// @Produce json
// @Param id path int true "Truck ID"
// @Success 200 {object} map[string]interface{} "Truck details"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Truck not found"
// @Router /admin/trucks/{id} [get]
// @Security Bearer
func (h *Handler) GetTruck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrTruckInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	truck, err := h.TruckUsecase.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": truck})
}

// UpdateTruck mengupdate informasi armada.
// @Summary Update truck information
// @Description Update truck details
// @Tags Trucks
// @Accept json
// @Produce json
// @Param id path int true "Truck ID"
// @Param truck body model.Truck true "Updated truck details"
// @Success 200 {object} map[string]interface{} "Truck updated"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Truck not found"
// @Router /admin/trucks/{id} [put]
// @Security Bearer
func (h *Handler) UpdateTruck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrTruckInvalidID.Error()})
		return
	}

	var input model.Truck
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.TruckUsecase.Update(ctx, id, &input); err != nil {
		if err == usecase.ErrTruckInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Armada berhasil diupdate",
		"data":    input,
	})
}

// DeleteTruck medeaktifkan armada.
// @Summary Delete truck
// @Description Deactivate a truck from fleet
// @Tags Trucks
// @Produce json
// @Param id path int true "Truck ID"
// @Success 200 {object} map[string]string "Truck deleted"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Truck not found"
// @Router /admin/trucks/{id} [delete]
// @Security Bearer
func (h *Handler) DeleteTruck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrTruckInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.TruckUsecase.Deactivate(ctx, id); err != nil {
		if err == usecase.ErrTruckInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Armada berhasil dideaktifkan"})
}
