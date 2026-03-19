package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/izzahnin/jalur-berlian-backend/internal/model"
)

// RegisterUserRoutes - POST /admin/users (1 endpoint).
func (h *Handler) RegisterUserRoutes(r *gin.RouterGroup) {
	r.POST("/admin/users", h.CreateUser)
}

// CreateUser membuat user baru (admin/customer).
// @Summary Create new user
// @Description Create a new admin or customer user with username, password, and role
// @Tags Users
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "User creation data"
// @Success 201 {object} object "User created successfully"
// @Failure 400 {object} map[string]string "Invalid input or role must be admin or customer"
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

	user, err := h.AuthUsecase.CreateUser(ctx, &input)
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
