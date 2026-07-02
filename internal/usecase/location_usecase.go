package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
	"github.com/izzahnin/jalur-berlian-backend/pkg/database"
)

const locationThrottleTTL = 30 * time.Second

type LocationUsecase struct {
	repo     repository.LocationRepository
	tripRepo *repository.TripRepository
	redis    *database.RedisClient
}

func NewLocationUsecase(repo repository.LocationRepository, tripRepo *repository.TripRepository, redis *database.RedisClient) *LocationUsecase {
	return &LocationUsecase{repo: repo, tripRepo: tripRepo, redis: redis}
}

func (u *LocationUsecase) SaveLocation(ctx context.Context, tripID int, lat, lon float64, ts time.Time) error {
	if tripID <= 0 {
		return ErrLocationInvalidTripID
	}

	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return ErrLocationInvalidCoords
	}

	trip, err := u.tripRepo.GetByID(ctx, tripID)
	if err != nil || trip == nil {
		return ErrTripNotFound
	}

	if trip.Status == "pickup" {
		return ErrLocationTripNotInTransit
	}

	if trip.Status == "delivered" {
		return ErrLocationTripDelivered
	}

	// Throttle: cegah spam lokasi lebih dari 1x per 30 detik per trip
	throttleKey := fmt.Sprintf("trip:%d:location_throttle", tripID)
	set, err := u.redis.Client().SetNX(ctx, throttleKey, "1", locationThrottleTTL).Result()
	if err == nil && !set {
		return ErrLocationThrottled
	}

	return u.repo.SaveLocation(ctx, tripID, lat, lon, ts)
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
