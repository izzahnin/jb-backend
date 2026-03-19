package repository

import (
	"context"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

// AuditLogRepository menangani akses data audit log ke database PostgreSQL.
// Setiap create/update/delete operation akan dicatat untuk compliance, debugging, dan forensic.
type AuditLogRepository struct {
	db *sqlx.DB
}

// NewAuditLogRepository membuat instance baru dari AuditLogRepository.
func NewAuditLogRepository(db *sqlx.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create menyimpan audit log entry baru ke database.
// Melakukan INSERT ke tabel audit_logs dengan user_id, action, table_name, record_id, old_values, new_values.
// Audit log dibuat dalam transaction yang sama dengan mutation (untuk atomicity).
// Returns: error jika database error terjadi.
// Note: Audit log HARUS berhasil (jangan ignore error), karena critical untuk compliance.
func (r *AuditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	query := `INSERT INTO audit_logs (user_id, action, table_name, record_id, old_values, new_values, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, NOW())
	          RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, log.UserID, log.Action, log.TableName, log.RecordID, log.OldValues, log.NewValues).
		Scan(&log.ID, &log.CreatedAt)
}

// GetByRecordID mengambil daftar audit log untuk sebuah record tertentu.
// Berguna untuk melihat history perubahan satu record (mis. "siapa saja yg edit truck #5?").
// Returns: slice dari audit log objects, atau error.
func (r *AuditLogRepository) GetByRecordID(ctx context.Context, tableName string, recordID int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	query := `SELECT id, user_id, action, table_name, record_id, old_values, new_values, created_at
	          FROM audit_logs
	          WHERE table_name = $1 AND record_id = $2
	          ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &logs, query, tableName, recordID); err != nil {
		return nil, err
	}
	return logs, nil
}

// GetByUserID mengambil daftar audit log untuk sebuah user.
// Berguna untuk investigate "apa yang dilakukan user #3 hari ini?".
// Returns: slice dari audit log objects, atau error.
func (r *AuditLogRepository) GetByUserID(ctx context.Context, userID int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	query := `SELECT id, user_id, action, table_name, record_id, old_values, new_values, created_at
	          FROM audit_logs
	          WHERE user_id = $1
	          ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &logs, query, userID); err != nil {
		return nil, err
	}
	return logs, nil
}
