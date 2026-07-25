package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/model"
)

// RegisterUserRoutes mendaftarkan user management endpoints (super_admin only).
// Note: PATCH /admin/profile is registered separately in routes.go line 76 with adminRoutes
// so all authenticated admins can update their own profile, not just super_admin.
func (h *Handler) RegisterUserRoutes(r *gin.RouterGroup) {
	r.GET("/admin/users", h.ListUsers)
	r.POST("/admin/users", h.CreateUser)
	r.DELETE("/admin/users/:id", h.DeleteUser)
	r.PATCH("/admin/users/:id/password", h.ResetUserPassword)
}

// CreateUser membuat user baru (super_admin-only).
// Hanya super_admin yang bisa membuat akun admin baru (admin_sales atau admin_ops).
// Full_name adalah optional, jika kosong akan default ke username.
// Password harus minimal 6 karakter.
// @Summary Create new admin user (super_admin only)
// @Description Create a new internal admin user with username, password, and role. Only super_admin can access. Full_name defaults to username if not provided. Password must be at least 6 characters.
// @Tags User Management
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "User creation data: username (required), password (required, min 6 chars), full_name (optional), role (required: super_admin, admin_sales, or admin_ops), is_active (optional, default true)"
// @Success 201 {object} object "User created successfully with user object containing id, username, full_name, role, is_active, created_at"
// @Failure 400 {object} map[string]string "Invalid input: username required, password required, password must be at least 6 characters, role must be one of: super_admin, admin_sales, admin_ops"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - super_admin role required"
// @Failure 409 {object} map[string]string "Conflict - username already exists"
// @Failure 422 {object} map[string]string "Unprocessable entity - failed to create user"
// @Failure 500 {object} map[string]string "Internal server error - database error"
// @Router /admin/users [post]
// @Security Bearer
func (h *Handler) CreateUser(c *gin.Context) {
	var input model.CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := h.UserUsecase.CreateUser(ctx, &input)
	if err != nil {
		if err.Error() == "username already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "role must be one of: super_admin, admin_sales, admin_ops" {
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

// ListUsers mengambil daftar semua users (super_admin-only).
// Mengembalikan list semua admin users tanpa password hash (secure).
// @Summary List all users (super_admin only)
// @Description Get list of all users with their username, role, and active status (password_hash excluded). Only super_admin can access. Returns message, count, and data array.
// @Tags User Management
// @Produce json
// @Success 200 {object} map[string]interface{} "Success response with message, count, and data array of users (each user has id, username, full_name, role, is_active, created_at)"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 403 {object} map[string]string "Forbidden - super_admin role required"
// @Failure 500 {object} map[string]string "Internal server error - database error"
// @Router /admin/users [get]
// @Security Bearer
func (h *Handler) ListUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
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

// UpdateProfile memungkinkan admin update profile mereka sendiri (fullname dan/atau password).
// User_id diambil dari JWT token, sehingga setiap user hanya bisa update profile mereka sendiri.
// Minimal 1 field harus diisi (full_name atau password). Password harus minimal 6 karakter.
// @Summary Update own profile (all admin roles)
// @Description Update your own profile information (full_name and/or password). All authenticated admins can access. At least one field must be provided. Password must be at least 6 characters. User_id is extracted from JWT token.
// @Tags User Management
// @Accept json
// @Produce json
// @Param body body model.UpdateProfileRequest true "Profile update data: full_name (optional), password (optional, min 6 chars). At least one required."
// @Success 200 {object} map[string]interface{} "Profile updated successfully with message and updated user object (id, username, full_name, role, is_active, created_at)"
// @Failure 400 {object} map[string]string "Bad request: at least one field (full_name or password) must be provided, or password must be at least 6 characters"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid JWT token"
// @Failure 404 {object} map[string]string "Not found - user not found in database"
// @Failure 422 {object} map[string]string "Unprocessable entity - failed to update profile"
// @Failure 500 {object} map[string]string "Internal server error - database error"
// @Router /admin/profile [patch]
// @Security Bearer
func (h *Handler) UpdateProfile(c *gin.Context) {
	// Extract user_id dari JWT token (set oleh middleware.AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in token"})
		return
	}

	var input model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := h.UserUsecase.UpdateProfile(ctx, userID.(int), &input)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "at least one field (full_name or password) must be provided" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "password must be at least 6 characters" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile berhasil diperbarui",
		"data":    user,
	})
}

// ResetUserPassword mereset password user lain (super_admin only).
// @Summary Reset password user (super_admin only)
// @Tags User Management
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body object true "New password (min 6 chars)"
// @Success 200 {object} map[string]string "Password berhasil direset"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "User not found"
// @Router /admin/users/{id}/password [patch]
// @Security Bearer
func (h *Handler) ResetUserPassword(c *gin.Context) {
	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var input struct {
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON tidak valid"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.UserUsecase.ResetPassword(ctx, targetID, input.NewPassword); err != nil {
		switch err.Error() {
		case "password must be at least 6 characters":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil direset"})
}

// DeleteUser menonaktifkan user (soft delete) berdasarkan ID (super_admin only).
func (h *Handler) DeleteUser(c *gin.Context) {
	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	// Cegah self-delete
	callerID, _ := c.Get("user_id")
	if callerID.(int) == targetID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak dapat menonaktifkan akun sendiri"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.UserUsecase.DeactivateUser(ctx, targetID); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User berhasil dinonaktifkan"})
}
