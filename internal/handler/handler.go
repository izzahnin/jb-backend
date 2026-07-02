// Package handler berisi semua HTTP endpoint handlers.
// Handlers diorganisir berdasarkan domain: auth, truck, order, location, public.
package handler

import (
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
	"github.com/izzahnin/jalur-berlian-backend/internal/usecase"
	"github.com/izzahnin/jalur-berlian-backend/pkg/database"
)

// Handler adalah root handler yang menyimpan semua dependencies.
// Setiap domain-specific handler menerima pointer ke root handler.
type Handler struct {
	// Repositories
	TruckRepo    *repository.TruckRepository
	OrderRepo    *repository.OrderRepository
	TripRepo     *repository.TripRepository
	DriverRepo   *repository.DriverRepository
	CustomerRepo *repository.CustomerRepository
	AuditRepo    *repository.AuditLogRepository
	LocRepo      repository.LocationRepository
	UserRepo     *repository.UserRepository

	// Usecases
	TruckUsecase    *usecase.TruckUsecase
	OrderUsecase    *usecase.OrderUsecase
	TripUsecase     *usecase.TripUsecase
	LocationUsecase *usecase.LocationUsecase
	CustomerUsecase *usecase.CustomerUsecase
	DriverUsecase   *usecase.DriverUsecase
	AuthUsecase     *usecase.AuthUsecase
	UserUsecase     *usecase.UserUsecase

	// Config
	JWTSecret string
	Redis     *database.RedisClient
}

// NewHandler membuat instance baru root handler.
func NewHandler(
	truckRepo *repository.TruckRepository,
	orderRepo *repository.OrderRepository,
	tripRepo *repository.TripRepository,
	driverRepo *repository.DriverRepository,
	customerRepo *repository.CustomerRepository,
	auditRepo *repository.AuditLogRepository,
	locRepo repository.LocationRepository,
	userRepo *repository.UserRepository,
	truckUsecase *usecase.TruckUsecase,
	orderUsecase *usecase.OrderUsecase,
	tripUsecase *usecase.TripUsecase,
	locationUsecase *usecase.LocationUsecase,
	customerUsecase *usecase.CustomerUsecase,
	driverUsecase *usecase.DriverUsecase,
	authUsecase *usecase.AuthUsecase,
	userUsecase *usecase.UserUsecase,
	jwtSecret string,
	redis *database.RedisClient,
) *Handler {
	return &Handler{
		TruckRepo:       truckRepo,
		OrderRepo:       orderRepo,
		TripRepo:        tripRepo,
		DriverRepo:      driverRepo,
		CustomerRepo:    customerRepo,
		AuditRepo:       auditRepo,
		LocRepo:         locRepo,
		UserRepo:        userRepo,
		TruckUsecase:    truckUsecase,
		OrderUsecase:    orderUsecase,
		TripUsecase:     tripUsecase,
		LocationUsecase: locationUsecase,
		CustomerUsecase: customerUsecase,
		DriverUsecase:   driverUsecase,
		AuthUsecase:     authUsecase,
		UserUsecase:     userUsecase,
		JWTSecret:       jwtSecret,
		Redis:           redis,
	}
}
