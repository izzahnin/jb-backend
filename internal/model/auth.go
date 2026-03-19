package model

// User merepresentasikan pengguna admin dalam sistem.
// Struct ini disimpan di PostgreSQL tabel users.
type User struct {
	ID           int    `db:"id" json:"id"`                         // Primary key auto-increment
	Username     string `db:"username" json:"username"`             // Unique username untuk login
	PasswordHash string `db:"password_hash" json:"-"`               // bcrypt hash, jangan expose ke JSON
	Role         string `db:"role" json:"role"`                     // 'admin' atau 'customer'
	IsActive     bool   `db:"is_active" json:"is_active"`           // Untuk soft delete user
	CreatedAt    string `db:"created_at" json:"created_at"`         // Timestamp saat user dibuat
}

// AuditLog merepresentasikan catatan perubahan data di sistem.
// Setiap create/update/delete akan dicatat untuk compliance dan debugging.
type AuditLog struct {
	ID        int64  `db:"id" json:"id"`                 // Primary key auto-increment
	UserID    int    `db:"user_id" json:"user_id"`       // Siapa yang melakukan aksi
	Action    string `db:"action" json:"action"`         // 'CREATE', 'UPDATE', 'DELETE'
	TableName string `db:"table_name" json:"table_name"` // 'trucks', 'orders', 'users'
	RecordID  int    `db:"record_id" json:"record_id"`   // ID baris yang dimodifikasi
	OldValues string `db:"old_values" json:"old_values"` // JSON snapshot untuk UPDATE
	NewValues string `db:"new_values" json:"new_values"` // JSON snapshot untuk UPDATE/CREATE
	CreatedAt string `db:"created_at" json:"created_at"` // Timestamp kapan aksi terjadi
}

// LoginRequest adalah payload dari POST /auth/login
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // Username (wajib)
	Password string `json:"password" binding:"required"` // Password plain text (bcrypt di server)
}

// LoginResponse adalah respon sukses setelah login
type LoginResponse struct {
	Token     string `json:"token"`      // JWT token untuk Authorization header
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp kapan token expired
	User      User   `json:"user"`       // User yang login (untuk UI)
}

// RegisterRequest adalah payload dari POST /auth/register (public user registration)
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`    // Username (wajib, unique)
	Password string `json:"password" binding:"required"`    // Password plain text (minimal 6 char)
	Role     string `json:"role"`                            // 'customer' (default), can request 'admin'
}

// RegisterResponse adalah respon sukses setelah registrasi
type RegisterResponse struct {
	Token     string `json:"token"`      // JWT token untuk langsung bisa login
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp kapan token expired
	User      User   `json:"user"`       // User yang baru dibuat
}

// CreateUserRequest adalah payload dari POST /admin/users (admin membuat user)
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"` // Username (wajib, unique)
	Password string `json:"password" binding:"required"` // Password plain text
	Role     string `json:"role" binding:"required"`     // 'admin' atau 'customer' (admin pilih)
	IsActive bool   `json:"is_active"`                    // Default: true
}
