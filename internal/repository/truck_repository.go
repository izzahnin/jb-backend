package repository

import (
	"context"
	"database/sql"
	"errors"

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
// Melakukan INSERT ke tabel trucks dengan plate_number dan driver_name.
// Returns: error jika ada constraint violation (duplicate plate_number) atau database error.
func (r *TruckRepository) Create(ctx context.Context, t *model.Truck) error {
	query := `INSERT INTO trucks (plate_number, driver_name, is_active) 
						VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRowContext(ctx, query, t.PlateNumber, t.DriverName, t.IsActive).Scan(&t.ID)
}

// FetchAll mengambil seluruh daftar truck dari database (deprecated: use FetchAllWithPagination).
// Query akan mengembalikan semua truck dengan ORDER BY id DESC.
// Returns: slice dari truck objects, atau error jika query gagal.
func (r *TruckRepository) FetchAll(ctx context.Context) ([]model.Truck, error) {
	var trucks []model.Truck
	query := `SELECT id, plate_number, driver_name, is_active FROM trucks ORDER BY id DESC`
	
	if err := r.db.SelectContext(ctx, &trucks, query); err != nil {
		return nil, err
	}
	return trucks, nil
}

// FetchAllWithPagination mengambil daftar truck dengan pagination.
// Melakukan SELECT dengan LIMIT dan OFFSET untuk membatasi hasil.
// Returns: slice dari truck objects, atau error jika query gagal.
func (r *TruckRepository) FetchAllWithPagination(ctx context.Context, limit, offset int) ([]model.Truck, error) {
	var trucks []model.Truck
	query := `SELECT id, plate_number, driver_name, is_active FROM trucks ORDER BY id DESC LIMIT $1 OFFSET $2`
	
	if err := r.db.SelectContext(ctx, &trucks, query, limit, offset); err != nil {
		return nil, err
	}
	return trucks, nil
}

// Count mengabung total jumlah truck di database.
// Returns: jumlah total truck, atau error jika query gagal.
func (r *TruckRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trucks`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}
	return count, nil
}

// GetByID mengambil detail single truck berdasarkan ID.
// Melakukan SELECT dengan WHERE id = $1.
// Returns: pointer ke truck object, atau error jika truck tidak ditemukan atau query gagal.
func (r *TruckRepository) GetByID(ctx context.Context, id int) (*model.Truck, error) {
	t := &model.Truck{}
	query := `SELECT id, plate_number, driver_name, is_active FROM trucks WHERE id = $1`
	if err := r.db.GetContext(ctx, t, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("truck not found")
		}
		return nil, err
	}
	return t, nil
}

// Update mengubah informasi truck (driver_name, is_active) berdasarkan ID.
// Melakukan UPDATE ke tabel trucks dengan WHERE id = $2.
// Returns: error jika truck tidak ditemukan atau query gagal.
func (r *TruckRepository) Update(ctx context.Context, id int, t *model.Truck) error {
	query := `UPDATE trucks SET driver_name = $1, is_active = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, t.DriverName, t.IsActive, id)
	return err
}

// Delete melakukan soft delete truck dengan mengeset is_active = false berdasarkan ID.
// Soft delete digunakan agar tidak menghilangkan data historis.
// Returns: error jika truck tidak ditemukan atau query gagal.
func (r *TruckRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE trucks SET is_active = false WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
