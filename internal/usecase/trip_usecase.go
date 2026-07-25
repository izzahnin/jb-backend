package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
)

// TripUsecase orchestrates operational trip lifecycle and its side effects.
type TripUsecase struct {
	tripRepo   *repository.TripRepository
	orderRepo  *repository.OrderRepository
	truckRepo  *repository.TruckRepository
	driverRepo *repository.DriverRepository
	auditRepo  *repository.AuditLogRepository
}

func NewTripUsecase(
	tripRepo *repository.TripRepository,
	orderRepo *repository.OrderRepository,
	truckRepo *repository.TruckRepository,
	driverRepo *repository.DriverRepository,
	auditRepo *repository.AuditLogRepository,
) *TripUsecase {
	return &TripUsecase{
		tripRepo:   tripRepo,
		orderRepo:  orderRepo,
		truckRepo:  truckRepo,
		driverRepo: driverRepo,
		auditRepo:  auditRepo,
	}
}

func (u *TripUsecase) CreateTrip(ctx context.Context, trip *model.Trip, actorUserID int) error {
	if trip.OrderID <= 0 {
		return ErrOrderInvalidID
	}
	if trip.TruckID <= 0 {
		return ErrTruckInvalidID
	}
	if trip.DriverID <= 0 {
		return ErrDriverInvalidID
	}

	existingTrips, err := u.tripRepo.CountByOrderID(ctx, trip.OrderID)
	if err != nil {
		return err
	}
	if existingTrips > 0 {
		return ErrTripAlreadyExistsForOrder
	}

	order, err := u.orderRepo.GetByID(ctx, trip.OrderID)
	if err != nil {
		return ErrOrderNotFound
	}
	if order.Status == "completed" || order.Status == "cancelled" {
		return ErrOrderInvalidStatusTransition
	}

	truck, err := u.truckRepo.GetByID(ctx, trip.TruckID)
	if err != nil {
		return ErrTruckNotFound
	}
	if !truck.IsActive || truck.Status != "available" {
		return ErrTruckInactive
	}

	driver, err := u.driverRepo.GetByID(ctx, trip.DriverID)
	if err != nil {
		return ErrDriverNotFound
	}
	if !driver.IsActive || driver.Status != "available" {
		return ErrDriverInactive
	}

	trip.Status = "pickup"
	trip.CreatedBy = &actorUserID

	var auditUserID *int
	if u.auditRepo != nil {
		auditUserID = &actorUserID
	}
	if err := u.tripRepo.CreateWithAssignments(ctx, trip, order.Status == "pending", auditUserID); err != nil {
		return err
	}

	return nil
}

func (u *TripUsecase) StartTrip(ctx context.Context, tripID int, containerNumber, sealNumber string, actorUserID int) error {
	if tripID <= 0 {
		return ErrTripInvalidID
	}
	if containerNumber == "" {
		return ErrTripContainerRequired
	}
	if sealNumber == "" {
		return ErrTripSealRequired
	}

	trip, err := u.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return ErrTripNotFound
	}
	if trip.Status != "pickup" {
		return ErrTripInvalidStatusTransition
	}

	oldValue, _ := json.Marshal(trip)
	if err := u.tripRepo.UpdateDispatch(ctx, tripID, containerNumber, sealNumber, time.Now().UTC(), &actorUserID); err != nil {
		return err
	}
	trip.ContainerNumber = containerNumber
	trip.SealNumber = sealNumber
	trip.Status = "in_transit"
	now := time.Now().UTC()
	trip.StartTime = &now

	if u.auditRepo != nil {
		newValue, _ := json.Marshal(trip)
		_ = u.auditRepo.Create(ctx, &model.AuditLog{
			UserID:    actorUserID,
			Action:    "UPDATE",
			TableName: "trips",
			RecordID:  trip.ID,
			OldValues: string(oldValue),
			NewValues: string(newValue),
		})
	}

	return nil
}

func (u *TripUsecase) CompleteTrip(ctx context.Context, tripID int, actorUserID int) error {
	if tripID <= 0 {
		return ErrTripInvalidID
	}

	trip, err := u.tripRepo.GetByID(ctx, tripID)
	if err != nil {
		return ErrTripNotFound
	}
	if trip.Status != "in_transit" {
		return ErrTripInvalidStatusTransition
	}

	oldValue, _ := json.Marshal(trip)
	now := time.Now().UTC()
	var auditUserID *int
	if u.auditRepo != nil {
		auditUserID = &actorUserID
	}
	if err := u.tripRepo.CompleteWithRelease(ctx, trip, now, &actorUserID, string(oldValue), auditUserID); err != nil {
		return err
	}

	return nil
}

func (u *TripUsecase) GetByOrderID(ctx context.Context, orderID int) (*model.Trip, error) {
	if orderID <= 0 {
		return nil, ErrOrderInvalidID
	}
	return u.tripRepo.GetByOrderID(ctx, orderID)
}

func (u *TripUsecase) GetByID(ctx context.Context, id int) (*model.Trip, error) {
	if id <= 0 {
		return nil, ErrTripInvalidID
	}
	return u.tripRepo.GetByID(ctx, id)
}

func (u *TripUsecase) GetAll(ctx context.Context) ([]model.Trip, error) {
	return u.tripRepo.FetchAll(ctx)
}
