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

// RegisterCustomerRoutes mendaftarkan CRUD endpoints untuk customer (semua admin).
func (h *Handler) RegisterCustomerRoutes(r *gin.RouterGroup) {
	r.GET("/admin/customers", h.ListCustomers)
	r.POST("/admin/customers", h.CreateCustomer)
	r.GET("/admin/customers/:id", h.GetCustomer)
	r.PATCH("/admin/customers/:id", h.UpdateCustomer)
	r.DELETE("/admin/customers/:id", h.DeleteCustomer)
}

// CreateCustomer membuat customer baru (all admin roles).
// @Summary Create customer
// @Description Create a new business customer (company/PIC/contact). All authenticated admins can access.
// @Tags Customers
// @Accept json
// @Produce json
// @Param body body model.CreateCustomerRequest true "Customer data: company_name, pic_name, phone, email, address, npwp"
// @Success 201 {object} map[string]interface{} "Customer created successfully"
// @Failure 400 {object} map[string]string "Bad request - missing required fields"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/customers [post]
// @Security Bearer
func (h *Handler) CreateCustomer(c *gin.Context) {
	var input model.CreateCustomerRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	customer, err := h.CustomerUsecase.Create(ctx, &input)
	if err != nil {
		switch err {
		case usecase.ErrCustomerCompanyRequired, usecase.ErrCustomerPICRequired, usecase.ErrCustomerPhoneRequired:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Customer berhasil dibuat", "data": customer})
}

// ListCustomers mengambil daftar semua customer.
// @Summary List customers
// @Description Get list of all customers. All authenticated admins can access.
// @Tags Customers
// @Produce json
// @Success 200 {object} map[string]interface{} "Customer list with count"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/customers [get]
// @Security Bearer
func (h *Handler) ListCustomers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	customers, err := h.CustomerUsecase.List(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": len(customers), "data": customers})
}

// GetCustomer mengambil detail customer berdasarkan ID.
// @Summary Get customer detail
// @Description Retrieve a customer by ID.
// @Tags Customers
// @Produce json
// @Param id path int true "Customer ID"
// @Success 200 {object} map[string]interface{} "Customer detail"
// @Failure 400 {object} map[string]string "Bad request - invalid customer ID"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 404 {object} map[string]string "Not found - customer tidak ditemukan"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/customers/{id} [get]
// @Security Bearer
func (h *Handler) GetCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrCustomerInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	customer, err := h.CustomerUsecase.GetByID(ctx, id)
	if err != nil {
		if err == usecase.ErrCustomerInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "customer not found" || err == usecase.ErrCustomerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrCustomerNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customer})
}

// UpdateCustomer memperbarui data customer (partial update).
// @Summary Update customer
// @Description Update customer fields (partial). All authenticated admins can access.
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path int true "Customer ID"
// @Param body body model.UpdateCustomerRequest true "Customer update data (partial)"
// @Success 200 {object} map[string]interface{} "Customer updated successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid data"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 404 {object} map[string]string "Not found - customer tidak ditemukan"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/customers/{id} [patch]
// @Security Bearer
func (h *Handler) UpdateCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrCustomerInvalidID.Error()})
		return
	}

	var input model.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	customer, err := h.CustomerUsecase.Update(ctx, id, &input)
	if err != nil {
		switch err {
		case usecase.ErrCustomerInvalidID, usecase.ErrCustomerCompanyRequired, usecase.ErrCustomerPICRequired, usecase.ErrCustomerPhoneRequired:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case usecase.ErrCustomerNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer berhasil diperbarui", "data": customer})
}

// DeleteCustomer menghapus customer berdasarkan ID.
// @Summary Delete customer
// @Description Delete a customer by ID. All authenticated admins can access.
// @Tags Customers
// @Produce json
// @Param id path int true "Customer ID"
// @Success 200 {object} map[string]string "Customer deleted successfully"
// @Failure 400 {object} map[string]string "Bad request - invalid customer ID"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 404 {object} map[string]string "Not found - customer tidak ditemukan"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/customers/{id} [delete]
// @Security Bearer
func (h *Handler) DeleteCustomer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": usecase.ErrCustomerInvalidID.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.CustomerUsecase.Delete(ctx, id); err != nil {
		if err == usecase.ErrCustomerInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "customer not found" || err == usecase.ErrCustomerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": usecase.ErrCustomerNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer berhasil dihapus"})
}
