package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

// TripRepository manages trip (surat jalan) persistence.
type TripRepository struct {
	db *sqlx.DB
}

func NewTripRepository(db *sqlx.DB) *TripRepository {
	return &TripRepository{db: db}
}

func (r *TripRepository) Create(ctx context.Context, t *model.Trip) error {
	query := `INSERT INTO trips (
		order_id,
		truck_id,
		driver_id,
		container_number,
		seal_number,
		status,
		is_active,
		start_time,
		end_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, trip_number, created_at`

	return r.db.QueryRowContext(ctx, query,
		t.OrderID,
		t.TruckID,
		t.DriverID,
		t.ContainerNumber,
		t.SealNumber,
		t.Status,
		true,
		t.StartTime,
		t.EndTime,
	).Scan(&t.ID, &t.TripNumber, &t.CreatedAt)
}

func (r *TripRepository) GetByID(ctx context.Context, id int) (*model.Trip, error) {
	var t model.Trip
	query := `SELECT id, order_id, truck_id, driver_id, trip_number, container_number, seal_number, status, is_active, start_time, end_time, created_at
	          FROM trips
	          WHERE id = $1 AND is_active = true`
	if err := r.db.GetContext(ctx, &t, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("trip not found")
		}
		return nil, err
	}
	return &t, nil
}

func (r *TripRepository) GetByOrderID(ctx context.Context, orderID int) (*model.Trip, error) {
	var t model.Trip
	query := `SELECT id, order_id, truck_id, driver_id, trip_number, container_number, seal_number, status, is_active, start_time, end_time, created_at
	          FROM trips
	          WHERE order_id = $1 AND is_active = true
	          ORDER BY created_at DESC
	          LIMIT 1`
	if err := r.db.GetContext(ctx, &t, query, orderID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *TripRepository) FetchByOrderID(ctx context.Context, orderID int) ([]model.Trip, error) {
	var trips []model.Trip
	query := `SELECT id, order_id, truck_id, driver_id, trip_number, container_number, seal_number, status, is_active, start_time, end_time, created_at
	          FROM trips
	          WHERE order_id = $1 AND is_active = true
	          ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &trips, query, orderID); err != nil {
		return nil, err
	}
	return trips, nil
}

func (r *TripRepository) FetchAll(ctx context.Context) ([]model.Trip, error) {
	var trips []model.Trip
	query := `SELECT id, order_id, truck_id, driver_id, trip_number, container_number, seal_number, status, is_active, start_time, end_time, created_at
	          FROM trips
	          WHERE is_active = true
	          ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &trips, query); err != nil {
		return nil, err
	}
	return trips, nil
}

func (r *TripRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE trips SET status = $1 WHERE id = $2 AND is_active = true`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *TripRepository) UpdateDispatch(ctx context.Context, id int, containerNumber, sealNumber string, startTime time.Time) error {
	query := `UPDATE trips
	          SET container_number = $1,
	              seal_number = $2,
	              start_time = $3,
	              status = 'in_transit'
	          WHERE id = $4 AND is_active = true`
	_, err := r.db.ExecContext(ctx, query, containerNumber, sealNumber, startTime, id)
	return err
}

func (r *TripRepository) MarkDelivered(ctx context.Context, id int, endTime time.Time) error {
	query := `UPDATE trips
	          SET end_time = $1,
	              status = 'delivered'
	          WHERE id = $2 AND is_active = true`
	_, err := r.db.ExecContext(ctx, query, endTime, id)
	return err
}

func (r *TripRepository) CountActiveByOrderID(ctx context.Context, orderID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trips WHERE order_id = $1 AND is_active = true AND status IN ('pickup', 'in_transit')`
	if err := r.db.GetContext(ctx, &count, query, orderID); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TripRepository) CountByOrderID(ctx context.Context, orderID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trips WHERE order_id = $1 AND is_active = true`
	if err := r.db.GetContext(ctx, &count, query, orderID); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TripRepository) CountActiveByTruckID(ctx context.Context, truckID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trips WHERE truck_id = $1 AND is_active = true AND status IN ('pickup', 'in_transit')`
	if err := r.db.GetContext(ctx, &count, query, truckID); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TripRepository) CountByOrderIDAndStatus(ctx context.Context, orderID int, status string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trips WHERE order_id = $1 AND is_active = true AND status = $2`
	if err := r.db.GetContext(ctx, &count, query, orderID, status); err != nil {
		return 0, err
	}
	return count, nil
}
