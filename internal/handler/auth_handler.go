package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/model"
)

// RegisterAuthRoutes mendaftarkan authentication endpoints (public - no auth required).
// POST /auth/login   - Login endpoint untuk generate JWT token
// POST /auth/logout  - Logout endpoint (stateless JWT)
func (h *Handler) RegisterAuthRoutes(r *gin.Engine) {
	r.POST("/auth/login", h.Login)
	r.POST("/auth/logout", h.Logout)
}

// RegisterSetupRoutes mendaftarkan initial setup endpoint (public - one-time only).
// POST /admin/setup - Create first super_admin account
func (h *Handler) RegisterSetupRoutes(r *gin.Engine) {
	r.POST("/admin/setup", h.AdminSetup)
}

// Login godoc
// @Summary User login - Get JWT token
// @Description Login with username and password to receive JWT token. Token is valid for 24 hours. Use token in Authorization: Bearer header for subsequent requests.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login credentials: username (required), password (required)"
// @Success 200 {object} model.LoginResponse "Login successful - returns token, expires_at timestamp, and user object"
// @Failure 400 {object} map[string]string "Bad request - invalid request format or missing username/password"
// @Failure 401 {object} map[string]string "Unauthorized - invalid username or password, or user inactive"
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var loginReq model.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	loginResp, err := h.AuthUsecase.Login(ctx, &loginReq)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, loginResp)
}

// AdminSetup godoc
// @Summary Initialize first admin user (super_admin)
// @Description Create the first super_admin user in the system (one-time endpoint, auto-disabled after first use). Returns JWT token for immediate use. Only succeeds if no admin exists yet.
// @Tags User Management
// @Accept json
// @Produce json
// @Param request body model.AdminSetupRequest true "Admin setup credentials: username, password (min 6 chars), optional full_name"
// @Success 201 {object} model.LoginResponse "Super admin created successfully with JWT token. User object includes immutable id and auto-set created_at"
// @Failure 400 {object} map[string]string "Invalid request format or invalid credentials"
// @Failure 409 {object} map[string]string "Conflict: Admin user already exists"
// @Router /admin/setup [post]
func (h *Handler) AdminSetup(c *gin.Context) {
	var setupReq model.AdminSetupRequest
	if err := c.ShouldBindJSON(&setupReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	loginResp, err := h.AuthUsecase.AdminSetup(ctx, &setupReq)
	if err != nil {
		if err.Error() == "admin user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, loginResp)
}

// Logout godoc
// @Summary User logout - Discard JWT token
// @Description Logout user (stateless JWT - server-side confirms, client discards token).
// @Tags Authentication
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]string "Logout successful - delete token from client storage"
// @Failure 401 {object} map[string]string "Unauthorized - missing Authorization header"
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header diperlukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout berhasil. Silahkan hapus token dari environment.",
	})
}
