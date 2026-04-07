package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/model"
)

// RegisterUserRoutes mendaftarkan user management endpoints.
func (h *Handler) RegisterUserRoutes(r *gin.RouterGroup) {
	r.GET("/admin/users", h.ListUsers)
	r.POST("/admin/users", h.CreateUser)
}

// CreateUser membuat user baru (admin-only).
// @Summary Create new user
// @Description Create a new admin or customer user with username, password, and role
// @Tags User Management
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "User creation data"
// @Success 201 {object} object "User created successfully"
// @Failure 400 {object} map[string]string "Invalid input or role must be admin or customer"
// @Failure 401 {object} map[string]string "Unauthorized or admin role required"
// @Failure 409 {object} map[string]string "Username already exists"
// @Failure 422 {object} map[string]string "Unprocessable entity"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/users [post]
// @Security Bearer
func (h *Handler) CreateUser(c *gin.Context) {
	var input model.CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := h.UserUsecase.CreateUser(ctx, &input)
	if err != nil {
		if err.Error() == "username already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "role must be 'admin' or 'customer'" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User berhasil dibuat",
		"data":    user,
	})
}

// ListUsers mengambil daftar semua users (admin-only).
// @Summary List all users
// @Description Get list of all users with their username, role, and active status (admin only)
// @Tags User Management
// @Produce json
// @Success 200 {object} map[string]interface{} "List of all users"
// @Failure 401 {object} map[string]string "Unauthorized or admin role required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /admin/users [get]
// @Security Bearer
func (h *Handler) ListUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := h.UserUsecase.ListUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Users retrieved successfully",
		"count":   len(users),
		"data":    users,
	})
}
