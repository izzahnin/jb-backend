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

func (u *LocationUsecase) SaveLocation(ctx context.Context, tripID int, lat, lon float64, speed *float64, ts time.Time) error {
	if tripID <= 0 {
		return ErrLocationInvalidTripID
	}

	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return ErrLocationInvalidCoords
	}

	return u.repo.SaveLocation(ctx, tripID, lat, lon, speed, ts)
}

func (u *LocationUsecase) GetLatest(ctx context.Context, tripID int) (*model.Location, error) {
	if tripID <= 0 {
		return nil, ErrLocationInvalidTripID
	}
	return u.repo.GetLatestLocation(ctx, tripID)
}

func (u *LocationUsecase) GetHistory(ctx context.Context, tripID int, limit int) ([]model.Location, error) {
	if tripID <= 0 {
		return nil, ErrLocationInvalidTripID
	}
	if limit <= 0 {
		limit = 50
	}
	return u.repo.GetHistory(ctx, tripID, limit)
}
