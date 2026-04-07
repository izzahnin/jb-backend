package usecase

import (
	"context"
	"errors"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
)

// UserUsecase menghandle business logic untuk user management.
// Bertanggung jawab untuk pembuatan user dan daftar user (admin-protected).
type UserUsecase struct {
	userRepo *repository.UserRepository
}

// NewUserUsecase membuat instance baru dari UserUsecase.
func NewUserUsecase(userRepo *repository.UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
	}
}

// CreateUser menangani pembuatan user baru oleh admin (admin-protected endpoint).
// Flow:
// 1. Validasi input (username, password, role)
// 2. Validasi role adalah 'admin' atau 'customer'
// 3. Hash password dengan bcrypt
// 4. Simpan user ke database
// 5. Return user info (tidak return token - admin yang beri password ke user)
// Returns: User model atau error jika validation/db gagal.
func (u *UserUsecase) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	// 1. VALIDATE: Username wajib diisi dan non-empty
	if req.Username == "" {
		return nil, errors.New("username required")
	}

	// 2. VALIDATE: Password wajib diisi dan minimal 6 karakter
	if req.Password == "" {
		return nil, errors.New("password required")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// 3. VALIDATE ROLE: Role harus salah satu dari 'admin' atau 'customer'
	if req.Role != "admin" && req.Role != "customer" {
		return nil, errors.New("role must be 'admin' or 'customer'")
	}

	// 4. CHECK DUPLICATE: Cek username sudah ada atau belum
	existingUser, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		// Username sudah ada
		return nil, errors.New("username already exists")
	}

	// 5. HASH PASSWORD: Generate bcrypt hash
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 6. SET DEFAULT is_active: Default true jika tidak dikirim, gunakan value jika dikirim
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// 7. CREATE USER: Simpan user baru ke database dengan role yang dipilih admin
	newUser := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Role:         req.Role,
		IsActive:     isActive,
	}

	err = u.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// 8. RESPONSE: Return user info tanpa password hash
	// Client (admin) akan memberi username+password ke user via secure channel
	return &model.User{
		ID:       newUser.ID,
		Username: newUser.Username,
		Role:     newUser.Role,
		IsActive: newUser.IsActive,
	}, nil
}

// ListUsers mengambil semua user dari database (admin-protected).
// Digunakan untuk admin melihat daftar semua users dan role mereka.
// Returns: slice of User objects (tanpa password hash) atau error jika database error.
func (u *UserUsecase) ListUsers(ctx context.Context) ([]*model.User, error) {
	users, err := u.userRepo.GetAll(ctx)
	if err != nil {
		return nil, errors.New("failed to fetch users")
	}

	// Filter out password_hash sebelum return (security)
	for _, user := range users {
		user.PasswordHash = "" // Don't expose password hash
	}

	return users, nil
}
