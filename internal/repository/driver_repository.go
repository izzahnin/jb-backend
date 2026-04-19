package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

// DriverRepository provides CRUD operations for drivers table.
type DriverRepository struct {
	db *sqlx.DB
}

func NewDriverRepository(db *sqlx.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

func (r *DriverRepository) Create(ctx context.Context, d *model.Driver) error {
	query := `INSERT INTO drivers (name, license_number, phone, status, is_active)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		d.Name, d.LicenseNumber, d.Phone, d.Status, d.IsActive,
	).Scan(&d.ID)
}

func (r *DriverRepository) FetchAll(ctx context.Context) ([]model.Driver, error) {
	var drivers []model.Driver
	query := `SELECT id, name, license_number, phone, status, is_active
	          FROM drivers
	          ORDER BY id DESC`
	if err := r.db.SelectContext(ctx, &drivers, query); err != nil {
		return nil, err
	}
	return drivers, nil
}

func (r *DriverRepository) GetByID(ctx context.Context, id int) (*model.Driver, error) {
	var d model.Driver
	query := `SELECT id, name, license_number, phone, status, is_active
	          FROM drivers
	          WHERE id = $1`
	if err := r.db.GetContext(ctx, &d, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("driver not found")
		}
		return nil, err
	}
	return &d, nil
}

func (r *DriverRepository) Update(ctx context.Context, id int, d *model.Driver) error {
	query := `UPDATE drivers
	          SET name = $1,
	              license_number = $2,
	              phone = $3,
	              status = $4,
	              is_active = $5
	          WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query,
		d.Name, d.LicenseNumber, d.Phone, d.Status, d.IsActive, id,
	)
	return err
}

func (r *DriverRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE drivers SET is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *DriverRepository) SetStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE drivers SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
