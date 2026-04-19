package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type LocationRepository interface {
	SaveLocation(ctx context.Context, tripID int, lat, lon float64, speed *float64, ts time.Time) error
	GetLatestLocation(ctx context.Context, tripID int) (*model.Location, error)
	GetHistory(ctx context.Context, tripID int, limit int) ([]model.Location, error)
}

type postgresLocationRepo struct {
	db *sqlx.DB
}

func NewPostgresLocationRepo(db *sqlx.DB) LocationRepository {
	return &postgresLocationRepo{db: db}
}

func (r *postgresLocationRepo) SaveLocation(ctx context.Context, tripID int, lat, lon float64, speed *float64, ts time.Time) error {
	query := `INSERT INTO locations (trip_id, latitude, longitude, speed, created_at)
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, tripID, lat, lon, speed, ts)
	return err
}

func (r *postgresLocationRepo) GetLatestLocation(ctx context.Context, tripID int) (*model.Location, error) {
	var loc model.Location
	query := `SELECT id, trip_id, latitude, longitude, speed, created_at
	          FROM locations
	          WHERE trip_id = $1
	          ORDER BY created_at DESC
	          LIMIT 1`
	if err := r.db.GetContext(ctx, &loc, query, tripID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &loc, nil
}

func (r *postgresLocationRepo) GetHistory(ctx context.Context, tripID int, limit int) ([]model.Location, error) {
	if limit <= 0 {
		limit = 50
	}

	var locations []model.Location
	query := `SELECT id, trip_id, latitude, longitude, speed, created_at
	          FROM locations
	          WHERE trip_id = $1
	          ORDER BY created_at DESC
	          LIMIT $2`
	if err := r.db.SelectContext(ctx, &locations, query, tripID, limit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.Location{}, nil
		}
		return nil, err
	}
	return locations, nil
}
