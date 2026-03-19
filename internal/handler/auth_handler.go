package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/model"
)

// RegisterAuthRoutes mendaftarkan semua auth endpoints (public, no auth required).
func (h *Handler) RegisterAuthRoutes(r *gin.Engine) {
	// 1. POST /auth/login - Login endpoint untuk generate JWT token
	r.POST("/auth/login", h.Login)

	// 2. POST /auth/register - Register endpoint untuk user membuat akun baru
	r.POST("/auth/register", h.Register)

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

// Register godoc
// @Summary User registration
// @Description Register new admin user with username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Registration credentials"
// @Success 201 {object} model.RegisterResponse "Registration successful"
// @Failure 400 {object} map[string]string "Invalid request format"
// @Failure 409 {object} map[string]string "Username already exists"
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var registerReq model.RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	registerResp, err := h.AuthUsecase.Register(ctx, &registerReq)
	if err != nil {
		if err.Error() == "username already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, registerResp)
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
