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

// Create menyimpan data truck baru ke database.
// Melakukan INSERT ke tabel trucks dengan plate_number, truck_type, status, is_active, dan created_at.
// Returns: error jika ada constraint violation (duplicate plate_number) atau database error.
func (r *TruckRepository) Create(ctx context.Context, t *model.Truck) error {
	query := `INSERT INTO trucks (plate_number, truck_type, status, is_active, created_at)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, t.PlateNumber, t.TruckType, t.Status, t.IsActive, time.Now()).Scan(&t.ID, &t.CreatedAt)
}

// FetchAll mengambil seluruh daftar truck dari database (deprecated: use FetchAllWithPagination).
// Query akan mengembalikan semua truck dengan ORDER BY created_at DESC (terbaru dulu).
// Returns: slice dari truck objects, atau error jika query gagal.
func (r *TruckRepository) FetchAll(ctx context.Context) ([]model.Truck, error) {
	var trucks []model.Truck
	query := `SELECT id, plate_number, truck_type, status, is_active, created_at FROM trucks WHERE is_active = true ORDER BY created_at DESC`
	
	if err := r.db.SelectContext(ctx, &trucks, query); err != nil {
		return nil, err
	}
	return trucks, nil
}

// FetchAllWithPagination mengambil daftar truck dengan pagination.
// Melakukan SELECT dengan LIMIT dan OFFSET untuk membatasi hasil, order by created_at DESC.
// Returns: slice dari truck objects, atau error jika query gagal.
func (r *TruckRepository) FetchAllWithPagination(ctx context.Context, limit, offset int) ([]model.Truck, error) {
	var trucks []model.Truck
	query := `SELECT id, plate_number, truck_type, status, is_active, created_at FROM trucks WHERE is_active = true ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	
	if err := r.db.SelectContext(ctx, &trucks, query, limit, offset); err != nil {
		return nil, err
	}
	return trucks, nil
}

// Count mengabung total jumlah truck di database.
// Returns: jumlah total truck, atau error jika query gagal.
func (r *TruckRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trucks WHERE is_active = true`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}
	return count, nil
}

// CountActive menghitung jumlah truck yang aktif (is_active = true).
// Digunakan untuk dashboard stats.
// Returns: jumlah truck aktif, atau error jika query gagal.
func (r *TruckRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trucks WHERE is_active = true`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}
	return count, nil
}

// GetByID mengambil detail single truck berdasarkan ID.
// Melakukan SELECT dengan WHERE id = $1, termasuk created_at.
// Returns: pointer ke truck object, atau error jika truck tidak ditemukan atau query gagal.
func (r *TruckRepository) GetByID(ctx context.Context, id int) (*model.Truck, error) {
	t := &model.Truck{}
	query := `SELECT id, plate_number, truck_type, status, is_active, created_at FROM trucks WHERE id = $1 AND is_active = true`
	if err := r.db.GetContext(ctx, t, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("truck not found")
		}
		return nil, err
	}
	return t, nil
}

// Update mengubah informasi truck (plate_number, driver_name, is_active) berdasarkan ID.
// Melakukan UPDATE ke tabel trucks dengan WHERE id = $4.
// Returns: error jika truck tidak ditemukan atau query gagal.
func (r *TruckRepository) Update(ctx context.Context, id int, t *model.Truck) error {
	query := `UPDATE trucks SET plate_number = $1, truck_type = $2, status = $3, is_active = $4 WHERE id = $5`
	_, err := r.db.ExecContext(ctx, query, t.PlateNumber, t.TruckType, t.Status, t.IsActive, id)
	return err
}

// Delete melakukan soft delete truck dengan mengeset is_active = false berdasarkan ID.
// Soft delete digunakan agar tidak menghilangkan data historis.
// Returns: error jika truck tidak ditemukan atau query gagal.
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
