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

// RegisterDriverRoutes mendaftarkan CRUD endpoints untuk driver (semua admin).
func (h *Handler) RegisterDriverRoutes(r *gin.RouterGroup) {
	r.GET("/admin/drivers", h.ListDrivers)
	r.POST("/admin/drivers", h.CreateDriver)
	r.GET("/admin/drivers/:id", h.GetDriver)
	r.PATCH("/admin/drivers/:id", h.UpdateDriver)
	r.DELETE("/admin/drivers/:id", h.DeleteDriver)
}

// CreateDriver membuat driver baru (all admin roles).
// @Summary Create driver
// @Description Create a new driver. All authenticated admins can access.
// @Tags Drivers
// @Accept json
// @Produce json
// @Param body body model.CreateDriverRequest true "Driver data: name, license_number, phone, status, is_active"
// @Success 201 {object} map[string]interface{} "Driver created successfully"
// @Failure 400 {object} map[string]string "Bad request - missing required fields"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/drivers [post]
// @Security Bearer
func (h *Handler) CreateDriver(c *gin.Context) {
	var input model.CreateDriverRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	userID, _ := c.Get("user_id")
	adminID := userID.(int)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	driver, err := h.DriverUsecase.Create(ctx, &input, &adminID)
	if err != nil {
		switch err {
		case usecase.ErrDriverNameRequired, usecase.ErrDriverLicenseRequired, usecase.ErrDriverPhoneRequired:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Driver berhasil dibuat", "data": driver})
}

// ListDrivers mengambil daftar semua driver.
// @Summary List drivers
// @Description Get list of all drivers. All authenticated admins can access.
// @Tags Drivers
// @Produce json
// @Success 200 {object} map[string]interface{} "Driver list with count"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/drivers [get]
// @Security Bearer
func (h *Handler) ListDrivers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	drivers, err := h.DriverUsecase.List(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": len(drivers), "data": drivers})
}

// GetDriver mengambil detail driver berdasarkan ID.
// @Summary Get driver detail
// @Description Retrieve a driver by ID.
// @Tags Drivers
// @Produce json
// @Param id path int true "Driver ID"
// @Success 200 {object} map[string]interface{} "Driver detail"
// @Failure 400 {object} map[string]string "Bad request - invalid driver ID"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 404 {object} map[string]string "Not found - driver tidak ditemukan"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/drivers/{id} [get]
// @Security Bearer
func (h *Handler) GetDriver(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrDriverInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	driver, err := h.DriverUsecase.GetByID(ctx, id)
	if err != nil {
		if err == usecase.ErrDriverInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "driver not found" || err == usecase.ErrDriverNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrDriverNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": driver})
}

// UpdateDriver memperbarui data driver (partial update).
// @Summary Update driver
// @Description Update driver fields (partial). All authenticated admins can access.
// @Tags Drivers
// @Accept json
// @Produce json
// @Param id path int true "Driver ID"
// @Param body body model.UpdateDriverRequest true "Driver update data (partial)"
// @Success 200 {object} map[string]interface{} "Driver updated successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid data"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 404 {object} map[string]string "Not found - driver tidak ditemukan"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/drivers/{id} [patch]
// @Security Bearer
func (h *Handler) UpdateDriver(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrDriverInvalidID.Error()})
		return
	}

	var input model.UpdateDriverRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	userID, _ := c.Get("user_id")
	adminID := userID.(int)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	driver, err := h.DriverUsecase.Update(ctx, id, &input, &adminID)
	if err != nil {
		switch err {
		case usecase.ErrDriverInvalidID, usecase.ErrDriverNameRequired, usecase.ErrDriverLicenseRequired, usecase.ErrDriverPhoneRequired:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrDriverNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case usecase.ErrDriverOnActiveTrip:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver berhasil diperbarui", "data": driver})
}

// DeleteDriver menghapus (soft delete) driver berdasarkan ID.
// @Summary Delete driver
// @Description Deactivate a driver by ID. All authenticated admins can access.
// @Tags Drivers
// @Produce json
// @Param id path int true "Driver ID"
// @Success 200 {object} map[string]string "Driver deleted successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid driver ID"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 404 {object} map[string]string "Not found - driver tidak ditemukan"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/drivers/{id} [delete]
// @Security Bearer
func (h *Handler) DeleteDriver(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrDriverInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.DriverUsecase.Delete(ctx, id); err != nil {
		if err == usecase.ErrDriverInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "driver not found" || err == usecase.ErrDriverNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrDriverNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver berhasil dihapus"})
}
