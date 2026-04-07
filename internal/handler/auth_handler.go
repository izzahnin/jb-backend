package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/model"
)

// RegisterAuthRoutes mendaftarkan semua auth endpoints.
func (h *Handler) RegisterAuthRoutes(r *gin.Engine) {
	// 1. POST /auth/login - Login endpoint untuk generate JWT token
	r.POST("/auth/login", h.Login)

	// 2. POST /admin/setup - Initialize admin pertama kali (one-time setup)
	r.POST("/admin/setup", h.AdminSetup)

	// 3. POST /auth/logout - Logout endpoint (stateless JWT)
	r.POST("/auth/logout", h.Logout)
}

// Login godoc
// @Summary User login 
// @Description Login with username and password to get JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login credentials"
// @Success 200 {object} model.LoginResponse "Login successful with JWT token"
// @Failure 400 {object} map[string]string "Invalid request format"
// @Failure 401 {object} map[string]string "Invalid credentials"
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
// @Summary Initialize first admin user
// @Description Create the first admin user in the system (one-time endpoint, auto-disabled after first use)
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.AdminSetupRequest true "Admin setup credentials"
// @Success 201 {object} model.LoginResponse "Admin created successfully with JWT token"
// @Failure 400 {object} map[string]string "Invalid request format"
// @Failure 409 {object} map[string]string "Admin already exists"
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
// @Summary User logout
// @Description Logout user (stateless JWT - just discard token on client side)
// @Tags Authentication
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 401 {object} map[string]string "Missing authorization header"
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
