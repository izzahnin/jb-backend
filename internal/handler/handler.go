// Package handler berisi semua HTTP endpoint handlers.
// Handlers diorganisir berdasarkan domain: auth, truck, order, location, public.
package handler

import (
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
	"github.com/izzahnin/jalur-berlian-backend/internal/usecase"
)

// Handler adalah root handler yang menyimpan semua dependencies.
// Setiap domain-specific handler menerima pointer ke root handler.
type Handler struct {
	// Repositories
	TruckRepo *repository.TruckRepository
	OrderRepo *repository.OrderRepository
	LocRepo   repository.LocationRepository
	UserRepo  *repository.UserRepository

	// Usecases
	TruckUsecase    *usecase.TruckUsecase
	OrderUsecase    *usecase.OrderUsecase
	LocationUsecase *usecase.LocationUsecase
	AuthUsecase     *usecase.AuthUsecase
	UserUsecase     *usecase.UserUsecase

	// Config
	JWTSecret string
}

// NewHandler membuat instance baru root handler.
func NewHandler(
	truckRepo *repository.TruckRepository,
	orderRepo *repository.OrderRepository,
	locRepo repository.LocationRepository,
	userRepo *repository.UserRepository,
	truckUsecase *usecase.TruckUsecase,
	orderUsecase *usecase.OrderUsecase,
	locationUsecase *usecase.LocationUsecase,
	authUsecase *usecase.AuthUsecase,
	userUsecase *usecase.UserUsecase,
	jwtSecret string,
) *Handler {
	return &Handler{
		TruckRepo:       truckRepo,
		OrderRepo:       orderRepo,
		LocRepo:         locRepo,
		UserRepo:        userRepo,
		TruckUsecase:    truckUsecase,
		OrderUsecase:    orderUsecase,
		LocationUsecase: locationUsecase,
		AuthUsecase:     authUsecase,
		UserUsecase:     userUsecase,
		JWTSecret:       jwtSecret,
	}
}
