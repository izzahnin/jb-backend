package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type TruckRepository struct {
	db *sqlx.DB
}

func NewTruckRepository(db *sqlx.DB) *TruckRepository {
	return &TruckRepository{db: db}
}

func (r *TruckRepository) Create(ctx context.Context, t *model.Truck) error {
	query := `INSERT INTO trucks (plate_number, truck_type, status, is_active, created_at, created_by)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, t.PlateNumber, t.TruckType, t.Status, t.IsActive, time.Now(), t.CreatedBy).Scan(&t.ID, &t.CreatedAt)
}

func (r *TruckRepository) FetchAll(ctx context.Context) ([]model.Truck, error) {
	var trucks []model.Truck
	query := `SELECT t.id, t.plate_number, t.truck_type, t.status, t.is_active, t.created_at, t.updated_at, t.created_by, t.updated_by,
	            u1.username AS created_by_name,
	            u2.username AS updated_by_name
	          FROM trucks t
	          LEFT JOIN users u1 ON t.created_by = u1.id
	          LEFT JOIN users u2 ON t.updated_by = u2.id
	          WHERE t.is_active = true ORDER BY t.created_at DESC`
	if err := r.db.SelectContext(ctx, &trucks, query); err != nil {
		return nil, err
	}
	return trucks, nil
}

func (r *TruckRepository) FetchAllWithPagination(ctx context.Context, limit, offset int) ([]model.Truck, error) {
	var trucks []model.Truck
	query := `SELECT t.id, t.plate_number, t.truck_type, t.status, t.is_active, t.created_at, t.updated_at, t.created_by, t.updated_by,
	            u1.username AS created_by_name,
	            u2.username AS updated_by_name
	          FROM trucks t
	          LEFT JOIN users u1 ON t.created_by = u1.id
	          LEFT JOIN users u2 ON t.updated_by = u2.id
	          WHERE t.is_active = true ORDER BY t.created_at DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &trucks, query, limit, offset); err != nil {
		return nil, err
	}
	return trucks, nil
}

func (r *TruckRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trucks WHERE is_active = true`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TruckRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trucks WHERE is_active = true`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TruckRepository) GetByID(ctx context.Context, id int) (*model.Truck, error) {
	t := &model.Truck{}
	query := `SELECT t.id, t.plate_number, t.truck_type, t.status, t.is_active, t.created_at, t.updated_at, t.created_by, t.updated_by,
	            u1.username AS created_by_name,
	            u2.username AS updated_by_name
	          FROM trucks t
	          LEFT JOIN users u1 ON t.created_by = u1.id
	          LEFT JOIN users u2 ON t.updated_by = u2.id
	          WHERE t.id = $1 AND t.is_active = true`
	if err := r.db.GetContext(ctx, t, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("truck not found")
		}
		return nil, err
	}
	return t, nil
}

func (r *TruckRepository) Update(ctx context.Context, id int, t *model.Truck) error {
	query := `UPDATE trucks SET plate_number = $1, truck_type = $2, status = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP, updated_by = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, t.PlateNumber, t.TruckType, t.Status, t.IsActive, t.UpdatedBy, id)
	return err
}

func (r *TruckRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE trucks SET is_active = false WHERE id = $1 AND is_active = true`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *TruckRepository) SetStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE trucks SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
