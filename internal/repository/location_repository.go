package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/pkg/database"
	"github.com/redis/go-redis/v9"
)

type LocationRepository interface {
	SaveLocation(ctx context.Context, truckID int, lat, lon float64, ts time.Time) error
	GetLatestLocation(ctx context.Context, truckID int) (*model.Location, error)
	GetHistory(ctx context.Context, truckID int, limit int) ([]model.Location, error)
}

type redisLocationRepo struct {
	rdb *database.RedisClient
}

func NewRedisLocationRepo(rdb *database.RedisClient) LocationRepository {
	return &redisLocationRepo{rdb: rdb}
}

func (r *redisLocationRepo) SaveLocation(ctx context.Context, truckID int, lat, lon float64, ts time.Time) error {
	c := r.rdb.Client()
	// GEOADD for latest position
	if err := c.GeoAdd(ctx, "trucks:geo", &redis.GeoLocation{
		Longitude: lon,
		Latitude:  lat,
		Name:      fmt.Sprintf("%d", truckID),
	}).Err(); err != nil {
		return err
	}

	// Push history as JSON and cap to last 1000 entries
	loc := model.Location{
		TruckID:   truckID,
		Latitude:  lat,
		Longitude: lon,
		CreatedAt: ts,
	}
	b, _ := json.Marshal(loc)
	key := fmt.Sprintf("trucks:%d:loc", truckID)
	if err := c.LPush(ctx, key, b).Err(); err != nil {
		return err
	}
	if err := c.LTrim(ctx, key, 0, 999).Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisLocationRepo) GetLatestLocation(ctx context.Context, truckID int) (*model.Location, error) {
	c := r.rdb.Client()
	key := fmt.Sprintf("trucks:%d:loc", truckID)
	raw, err := c.LIndex(ctx, key, 0).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var loc model.Location
	if err := json.Unmarshal([]byte(raw), &loc); err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *redisLocationRepo) GetHistory(ctx context.Context, truckID int, limit int) ([]model.Location, error) {
	c := r.rdb.Client()
	key := fmt.Sprintf("trucks:%d:loc", truckID)
	vals, err := c.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	res := make([]model.Location, 0, len(vals))
	for _, v := range vals {
		var l model.Location
		if err := json.Unmarshal([]byte(v), &l); err == nil {
			res = append(res, l)
		}
	}
	return res, nil
}
