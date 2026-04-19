package model

import "time"

// User merepresentasikan pengguna admin dalam sistem.
// Struct ini disimpan di PostgreSQL tabel users.
type User struct {
	ID           int       `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	FullName     string    `db:"full_name" json:"full_name"`
	Role         string    `db:"role" json:"role"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// AuditLog merepresentasikan catatan perubahan data di sistem.
// Setiap create/update/delete akan dicatat untuk compliance dan debugging.
type AuditLog struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Action    string    `db:"action" json:"action"`
	TableName string    `db:"table_name" json:"table_name"`
	RecordID  int       `db:"record_id" json:"record_id"`
	OldValues string    `db:"old_values" json:"old_values"`
	NewValues string    `db:"new_values" json:"new_values"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// LoginRequest adalah payload dari POST /auth/login
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"superadmin"` // Username (wajib)
	Password string `json:"password" binding:"required" example:"password123"` // Password plain text (bcrypt di server)
}

// LoginResponse adalah respon sukses setelah login
type LoginResponse struct {
	Token     string `json:"token"`      // JWT token untuk Authorization header
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp kapan token expired
	User      User   `json:"user"`       // User yang login (untuk UI)
}

// CreateUserRequest adalah payload dari POST /admin/users (super_admin membuat user baru)
// Hanya super_admin yang bisa membuat user baru via endpoint ini.
type CreateUserRequest struct {
	Username string `json:"username" binding:"required" example:"john.sales"` // Username (wajib, unique)
	FullName string `json:"full_name" example:"John Sales Manager"`                   // Full name untuk display (optional, default = username jika kosong)
	Password string `json:"password" binding:"required" example:"password123456"` // Password plain text (akan di-hash dengan bcrypt)
	Role     string `json:"role" binding:"required" example:"admin_sales"`     // Role: 'super_admin', 'admin_sales', atau 'admin_ops'
	IsActive *bool  `json:"is_active" example:"true"`                    // Pointer untuk default true jika tidak dikirim
}

// AdminSetupRequest adalah payload dari POST /admin/setup (initialize admin pertama kali)
// Endpoint ini hanya bisa dijalankan sekali (one-time setup).
// Setelah admin pertama dibuat, gunakan POST /admin/users untuk user admin lainnya.
type AdminSetupRequest struct {
	Username string `json:"username" binding:"required" example:"superadmin"` // Admin username (wajib, unique)
	Password string `json:"password" binding:"required" example:"admin123456"` // Admin password plain text (minimal 6 char)
	FullName string `json:"full_name" example:"System Administrator"`                   // Full name untuk display (optional, default = username jika kosong)
}

// UpdateProfileRequest adalah payload dari PATCH /admin/profile (admin update profile mereka)
// Bisa update fullname dan/atau password. Minimal satu field harus diisi.
type UpdateProfileRequest struct {
	FullName string `json:"full_name" example:"Bambang Suryanto"` // Update display name (optional)
	Password string `json:"password" example:"newpassword123456"`   // Update password (optional, minimal 6 karakter jika diisi)
}
