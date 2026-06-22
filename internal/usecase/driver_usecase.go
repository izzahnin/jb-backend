package usecase

import (
	"context"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
)

type DriverUsecase struct {
	repo     *repository.DriverRepository
	tripRepo *repository.TripRepository
}

func NewDriverUsecase(repo *repository.DriverRepository, tripRepo *repository.TripRepository) *DriverUsecase {
	return &DriverUsecase{repo: repo, tripRepo: tripRepo}
}

func (u *DriverUsecase) Create(ctx context.Context, req *model.CreateDriverRequest, createdBy *int) (*model.Driver, error) {
	if req.Name == "" {
		return nil, ErrDriverNameRequired
	}
	if req.LicenseNumber == "" {
		return nil, ErrDriverLicenseRequired
	}
	if req.Phone == "" {
		return nil, ErrDriverPhoneRequired
	}

	driver := &model.Driver{
		Name:          req.Name,
		LicenseNumber: req.LicenseNumber,
		Phone:         req.Phone,
		Status:        req.Status,
		IsActive:      req.IsActive,
		CreatedBy:     createdBy,
	}

	if driver.Status == "" {
		driver.Status = "available"
	}
	if !driver.IsActive {
		driver.IsActive = true
	}

	if err := u.repo.Create(ctx, driver); err != nil {
		return nil, err
	}

	return driver, nil
}

func (u *DriverUsecase) List(ctx context.Context) ([]model.Driver, error) {
	return u.repo.FetchAll(ctx)
}

func (u *DriverUsecase) GetByID(ctx context.Context, id int) (*model.Driver, error) {
	if id <= 0 {
		return nil, ErrDriverInvalidID
	}
	return u.repo.GetByID(ctx, id)
}

func (u *DriverUsecase) Update(ctx context.Context, id int, req *model.UpdateDriverRequest, updatedBy *int) (*model.Driver, error) {
	if id <= 0 {
		return nil, ErrDriverInvalidID
	}

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	originalStatus := existing.Status

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.LicenseNumber != nil {
		existing.LicenseNumber = *req.LicenseNumber
	}
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if existing.Name == "" {
		return nil, ErrDriverNameRequired
	}
	if existing.LicenseNumber == "" {
		return nil, ErrDriverLicenseRequired
	}
	if existing.Phone == "" {
		return nil, ErrDriverPhoneRequired
	}

	// Block status change away from on_duty if driver has active trip
	if originalStatus == "on_duty" && existing.Status != "on_duty" {
		count, err := u.tripRepo.CountActiveByDriverID(ctx, id)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrDriverOnActiveTrip
		}
	}

	existing.UpdatedBy = updatedBy

	if err := u.repo.Update(ctx, id, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (u *DriverUsecase) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrDriverInvalidID
	}
	return u.repo.Delete(ctx, id)
}
