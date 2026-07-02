package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

// UserRepository menangani akses data user ke database PostgreSQL.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository membuat instance baru dari UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create menyimpan user baru ke database.
// Melakukan INSERT ke tabel users dengan username, password_hash, role, is_active.
// Returns: error jika ada constraint violation (duplicate username) atau database error.
func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	query := `INSERT INTO users (username, password_hash, full_name, role, is_active, created_at) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, u.Username, u.PasswordHash, u.FullName, u.Role, u.IsActive, time.Now()).Scan(&u.ID, &u.CreatedAt)
}

// GetByUsername mengambil user berdasarkan username.
// Digunakan pada login flow untuk cari user dan validate password.
// Returns: pointer ke user object, atau error jika user tidak ditemukan.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, username, password_hash, full_name, role, is_active, created_at FROM users WHERE username = $1`
	if err := r.db.GetContext(ctx, u, query, username); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return u, nil
}

// GetByID mengambil user berdasarkan ID.
// Digunakan untuk audit trail (untuk translate user_id → username).
// Returns: pointer ke user object, atau error jika user tidak ditemukan.
func (r *UserRepository) GetByID(ctx context.Context, id int) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, username, password_hash, full_name, role, is_active, created_at FROM users WHERE id = $1`
	if err := r.db.GetContext(ctx, u, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return u, nil
}

// AdminExists mengecek apakah sudah ada admin user di sistem.
// Digunakan untuk endpoint /admin/setup agar hanya bisa dijalankan sekali.
// Returns: true jika ada minimal 1 admin user, false jika belum ada.
func (r *UserRepository) AdminExists(ctx context.Context) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE role = 'super_admin'`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count menghitung total jumlah user di database.
// Digunakan untuk dashboard stats.
// Returns: jumlah total user, atau error jika query gagal.
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}
	return count, nil
}

// CountByRole menghitung jumlah user berdasarkan role.
// Digunakan untuk dashboard stats breakdown (admin count).
// Returns: jumlah user dengan role tertentu, atau error jika query gagal.
func (r *UserRepository) CountByRole(ctx context.Context, role string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE role = $1`
	if err := r.db.GetContext(ctx, &count, query, role); err != nil {
		return 0, err
	}
	return count, nil
}

// GetAll mengambil semua user dari database.
// Digunakan untuk admin melihat daftar semua users.
// Returns: slice of User objects atau error jika database error.
func (r *UserRepository) GetAll(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	query := `SELECT id, username, password_hash, full_name, role, is_active, created_at FROM users ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &users, query); err != nil {
		return nil, err
	}
	return users, nil
}

// Update memperbarui informasi user yang sudah ada (fullname, password, is_active).
// Digunakan untuk admin update profile mereka atau super_admin update user lain.
// Returns: error jika update gagal atau user tidak ditemukan.
func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	query := `UPDATE users SET password_hash = $1, full_name = $2, is_active = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, u.PasswordHash, u.FullName, u.IsActive, u.ID)
	return err
}

// UpdatePasswordHash mengganti password hash user berdasarkan ID (super_admin only).
func (r *UserRepository) UpdatePasswordHash(ctx context.Context, id int, hash string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("user not found")
	}
	return nil
}

// Deactivate melakukan soft delete user dengan mengeset is_active = false.
func (r *UserRepository) Deactivate(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET is_active = false WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("user not found")
	}
	return nil
}
