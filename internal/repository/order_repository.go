package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db *sqlx.DB // connection pool ke database PostgreSQL
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder menyimpan data order baru ke database.
// Melakukan INSERT ke tabel orders dengan order_number, truck_id, origin, destination, status.
// Status akan di-set ke value yang dikirim dari usecase (biasanya "pending").
// Returns: error jika ada constraint violation (duplicate order_number) atau database error.
func (r *OrderRepository) CreateOrder(ctx context.Context, o *model.Order) error {
	query := `INSERT INTO orders (order_number, truck_id, origin, destination, status) 
						VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,o.OrderNumber, o.TruckID, o.Origin, o.Destination, o.Status).Scan(&o.ID, &o.CreatedAt)
}

// FetchAll mengambil seluruh daftar order dari database (deprecated: use FetchAllWithPagination).
// Query akan mengembalikan semua order dengan ORDER BY created_at DESC (terbaru dulu).
// Returns: slice dari order pointers, atau error jika query gagal.
func (r *OrderRepository) FetchAll(ctx context.Context) ([]*model.Order, error) {
	var orders []*model.Order
	query := `SELECT id, order_number, truck_id, status, origin, destination, created_at
						FROM orders ORDER BY created_at DESC`
	
	if err := r.db.SelectContext(ctx, &orders, query); err != nil {
		return nil, err
	}
	return orders, nil
}

// FetchAllWithPagination mengambil daftar order dengan pagination.
// Melakukan SELECT dengan LIMIT dan OFFSET untuk membatasi hasil.
// Returns: slice dari order pointers, atau error jika query gagal.
func (r *OrderRepository) FetchAllWithPagination(ctx context.Context, limit, offset int) ([]*model.Order, error) {
	var orders []*model.Order
	query := `SELECT id, order_number, truck_id, status, origin, destination, created_at
						FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	
	if err := r.db.SelectContext(ctx, &orders, query, limit, offset); err != nil {
		return nil, err
	}
	return orders, nil
}

// Count menghitung total jumlah order di database.
// Returns: jumlah total order, atau error jika query gagal.
func (r *OrderRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM orders`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}
	return count, nil
}

// GetByID mengambil detail single order berdasarkan ID.
// Melakukan SELECT dengan WHERE id = $1.
// Returns: pointer ke order object, atau error jika order tidak ditemukan atau query gagal.
func (r *OrderRepository) GetByID(ctx context.Context, id int) (*model.Order, error) {
	o := &model.Order{}
	query := `SELECT id, order_number, truck_id, status, origin, destination, created_at
						FROM orders WHERE id = $1`
	if err := r.db.GetContext(ctx, o, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return o, nil
}

// GetByOrderNumber mengambil detail order berdasarkan nomor order (order_number).
// Digunakan untuk customer tracking API yang tidak memerlukan order ID.
// Returns: pointer ke order object, atau error jika tidak ditemukan atau query gagal.
func (r *OrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	o := &model.Order{}
	query := `SELECT id, order_number, truck_id, status, origin, destination, created_at
						FROM orders WHERE order_number = $1`
	if err := r.db.GetContext(ctx, o, query, orderNumber); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return o, nil
}

// UpdateStatus mengubah status order berdasarkan ID.
// Status harus valid sesuai CHECK constraint di database.
// Returns: error jika order tidak ditemukan atau query gagal.
func (r *OrderRepository) UpdateStatus(ctx context.Context, id int, newStatus string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, newStatus, id)
	return err
}

// AssignTruck mengassign truck ke order dan mengubah status menjadi "in_transit".
// Melakukan UPDATE truck_id dan status menjadi "pickup" (truck assigned, waiting at warehouse).
	// Returns: error jika order tidak ditemukan atau query gagal.
	func (r *OrderRepository) AssignTruck(ctx context.Context, orderID int, truckID int) error {
		query := `UPDATE orders SET truck_id = $1, status = 'pickup' WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, truckID, orderID)
	return err
}

// Delete melakukan soft delete order dengan status "cancelled" atau hard delete berdasarkan context.
// Saat ini menggunakan hard delete (DELETE FROM) - dapat diubah ke soft delete jika diperlukan.
// Returns: error jika order tidak ditemukan atau query gagal.
func (r *OrderRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM orders WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
