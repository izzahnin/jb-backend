package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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
		end_time,
		created_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
		t.CreatedBy,
	).Scan(&t.ID, &t.TripNumber, &t.CreatedAt)
}

func (r *TripRepository) CreateWithAssignments(ctx context.Context, t *model.Trip, updateOrderStatus bool, auditUserID *int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO trips (
		order_id,
		truck_id,
		driver_id,
		container_number,
		seal_number,
		status,
		is_active,
		start_time,
		end_time,
		created_by
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id, trip_number, created_at`

	if err := tx.QueryRowContext(ctx, query,
		t.OrderID,
		t.TruckID,
		t.DriverID,
		t.ContainerNumber,
		t.SealNumber,
		t.Status,
		true,
		t.StartTime,
		t.EndTime,
		t.CreatedBy,
	).Scan(&t.ID, &t.TripNumber, &t.CreatedAt); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE trucks SET status = $1 WHERE id = $2`, "on_duty", t.TruckID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE drivers SET status = $1 WHERE id = $2`, "on_duty", t.DriverID); err != nil {
		return err
	}
	if updateOrderStatus {
		if _, err := tx.ExecContext(ctx, `UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND is_active = true`, "partial", t.OrderID); err != nil {
			return err
		}
	}

	if auditUserID != nil {
		newValue, err := json.Marshal(t)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO audit_logs (user_id, action, table_name, record_id, old_values, new_values, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
			*auditUserID, "CREATE", "trips", t.ID, "", string(newValue),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *TripRepository) GetByID(ctx context.Context, id int) (*model.Trip, error) {
	var t model.Trip
	query := `SELECT tr.id, tr.order_id, tr.truck_id, tr.driver_id, tr.trip_number, tr.container_number, tr.seal_number,
	            tr.status, tr.is_active, tr.start_time, tr.end_time, tr.created_at, tr.created_by, tr.started_by, tr.completed_by,
	            u1.username AS started_by_name,
	            u2.username AS completed_by_name,
	            tk.plate_number AS truck_plate_number,
	            tk.is_active AS truck_is_active,
	            d.name AS driver_name
	          FROM trips tr
	          LEFT JOIN users u1 ON tr.started_by = u1.id
	          LEFT JOIN users u2 ON tr.completed_by = u2.id
	          LEFT JOIN trucks tk ON tr.truck_id = tk.id
	          LEFT JOIN drivers d ON tr.driver_id = d.id
	          WHERE tr.id = $1 AND tr.is_active = true`
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
	query := `SELECT tr.id, tr.order_id, tr.truck_id, tr.driver_id, tr.trip_number, tr.container_number, tr.seal_number,
	            tr.status, tr.is_active, tr.start_time, tr.end_time, tr.created_at, tr.created_by, tr.started_by, tr.completed_by,
	            u1.username AS started_by_name,
	            u2.username AS completed_by_name,
	            tk.plate_number AS truck_plate_number,
	            tk.is_active AS truck_is_active,
	            d.name AS driver_name
	          FROM trips tr
	          LEFT JOIN users u1 ON tr.started_by = u1.id
	          LEFT JOIN users u2 ON tr.completed_by = u2.id
	          LEFT JOIN trucks tk ON tr.truck_id = tk.id
	          LEFT JOIN drivers d ON tr.driver_id = d.id
	          WHERE tr.order_id = $1 AND tr.is_active = true
	          ORDER BY tr.created_at DESC
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
	query := `SELECT tr.id, tr.order_id, tr.truck_id, tr.driver_id, tr.trip_number, tr.container_number, tr.seal_number,
	            tr.status, tr.is_active, tr.start_time, tr.end_time, tr.created_at, tr.created_by, tr.started_by, tr.completed_by,
	            u1.username AS started_by_name,
	            u2.username AS completed_by_name,
	            tk.plate_number AS truck_plate_number,
	            tk.is_active AS truck_is_active,
	            d.name AS driver_name
	          FROM trips tr
	          LEFT JOIN users u1 ON tr.started_by = u1.id
	          LEFT JOIN users u2 ON tr.completed_by = u2.id
	          LEFT JOIN trucks tk ON tr.truck_id = tk.id
	          LEFT JOIN drivers d ON tr.driver_id = d.id
	          WHERE tr.order_id = $1 AND tr.is_active = true
	          ORDER BY tr.created_at DESC`
	if err := r.db.SelectContext(ctx, &trips, query, orderID); err != nil {
		return nil, err
	}
	return trips, nil
}

func (r *TripRepository) FetchAll(ctx context.Context) ([]model.Trip, error) {
	var trips []model.Trip
	query := `SELECT tr.id, tr.order_id, tr.truck_id, tr.driver_id, tr.trip_number, tr.container_number, tr.seal_number,
	            tr.status, tr.is_active, tr.start_time, tr.end_time, tr.created_at, tr.created_by, tr.started_by, tr.completed_by,
	            u1.username AS started_by_name,
	            u2.username AS completed_by_name,
	            tk.plate_number AS truck_plate_number,
	            tk.is_active AS truck_is_active,
	            d.name AS driver_name
	          FROM trips tr
	          LEFT JOIN users u1 ON tr.started_by = u1.id
	          LEFT JOIN users u2 ON tr.completed_by = u2.id
	          LEFT JOIN trucks tk ON tr.truck_id = tk.id
	          LEFT JOIN drivers d ON tr.driver_id = d.id
	          WHERE tr.is_active = true
	          ORDER BY tr.created_at DESC`
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

func (r *TripRepository) UpdateDispatch(ctx context.Context, id int, containerNumber, sealNumber string, startTime time.Time, startedBy *int) error {
	query := `UPDATE trips
	          SET container_number = $1,
	              seal_number = $2,
	              start_time = $3,
	              status = 'in_transit',
	              started_by = $4
	          WHERE id = $5 AND is_active = true`
	_, err := r.db.ExecContext(ctx, query, containerNumber, sealNumber, startTime, startedBy, id)
	return err
}

func (r *TripRepository) MarkDelivered(ctx context.Context, id int, endTime time.Time, completedBy *int) error {
	query := `UPDATE trips
	          SET end_time = $1,
	              status = 'delivered',
	              completed_by = $2
	          WHERE id = $3 AND is_active = true`
	_, err := r.db.ExecContext(ctx, query, endTime, completedBy, id)
	return err
}

func (r *TripRepository) CompleteWithRelease(ctx context.Context, t *model.Trip, endTime time.Time, completedBy *int, oldValues string, auditUserID *int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`UPDATE trips
		  SET end_time = $1,
		      status = 'delivered',
		      completed_by = $2
		  WHERE id = $3 AND is_active = true`,
		endTime, completedBy, t.ID,
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE trucks SET status = $1 WHERE id = $2`, "available", t.TruckID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE drivers SET status = $1 WHERE id = $2`, "available", t.DriverID); err != nil {
		return err
	}

	var totalTrips int
	if err := tx.GetContext(ctx, &totalTrips, `SELECT COUNT(*) FROM trips WHERE order_id = $1 AND is_active = true`, t.OrderID); err != nil {
		return err
	}
	var deliveredTrips int
	if err := tx.GetContext(ctx, &deliveredTrips, `SELECT COUNT(*) FROM trips WHERE order_id = $1 AND is_active = true AND status = $2`, t.OrderID, "delivered"); err != nil {
		return err
	}

	orderStatus := "partial"
	if totalTrips > 0 && totalTrips == deliveredTrips {
		orderStatus = "completed"
	}
	if _, err := tx.ExecContext(ctx, `UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND is_active = true`, orderStatus, t.OrderID); err != nil {
		return err
	}

	t.Status = "delivered"
	t.EndTime = &endTime
	t.CompletedBy = completedBy
	if auditUserID != nil {
		newValue, err := json.Marshal(t)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO audit_logs (user_id, action, table_name, record_id, old_values, new_values, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
			*auditUserID, "UPDATE", "trips", t.ID, oldValues, string(newValue),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
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

func (r *TripRepository) CountActiveByDriverID(ctx context.Context, driverID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM trips WHERE driver_id = $1 AND is_active = true AND status IN ('pickup', 'in_transit')`
	if err := r.db.GetContext(ctx, &count, query, driverID); err != nil {
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
