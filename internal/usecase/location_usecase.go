package usecase

import (
	"context"
	"time"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
)

type LocationUsecase struct {
	repo repository.LocationRepository
}

func NewLocationUsecase(repo repository.LocationRepository) *LocationUsecase {
	return &LocationUsecase{repo: repo}
}

func (u *LocationUsecase) SaveLocation(ctx context.Context, truckID int, lat, lon float64, ts time.Time) error {
	// Basic validation
	if truckID <= 0 {
		return ErrTruckInvalidID
	}
	return u.repo.SaveLocation(ctx, truckID, lat, lon, ts)
}

func (u *LocationUsecase) GetLatest(ctx context.Context, truckID int) (*model.Location, error) {
	return u.repo.GetLatestLocation(ctx, truckID)
}

func (u *LocationUsecase) GetHistory(ctx context.Context, truckID int, limit int) ([]model.Location, error) {
	if limit <= 0 {
		limit = 50
	}
	return u.repo.GetHistory(ctx, truckID, limit)
}
