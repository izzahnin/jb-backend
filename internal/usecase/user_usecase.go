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

// CreateUser menangani pembuatan user baru oleh super_admin (super_admin-protected endpoint).
// Hanya super_admin yang bisa create user. Endpoint ini untuk membuat admin_sales atau admin_ops accounts.
// Flow:
// 1. Validasi input (username, password, role)
// 2. Validasi role adalah 'super_admin', 'admin_sales', atau 'admin_ops'
// 3. Hash password dengan bcrypt
// 4. Simpan user ke database
// 5. Return user info (tidak return token - super_admin yang share password dengan user baru)
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

	// 3. VALIDATE ROLE: Role harus salah satu role internal
	validRoles := map[string]bool{
		"super_admin": true,
		"admin_sales": true,
		"admin_ops":   true,
	}
	if !validRoles[req.Role] {
		return nil, errors.New("role must be one of: super_admin, admin_sales, admin_ops")
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

	// 7. SET DEFAULT FULLNAME: Jika fullname kosong, gunakan username sebagai display name
	fullName := req.FullName
	if fullName == "" {
		fullName = req.Username
	}

	// 8. CREATE USER: Simpan user baru ke database dengan role yang dipilih super_admin
	newUser := &model.User{
		Username:     req.Username,
		FullName:     fullName,
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
		FullName: newUser.FullName,
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

// UpdateProfile memungkinkan admin update profile mereka sendiri (fullname dan/atau password).
// Flow:
// 1. Validasi user exists
// 2. Jika ada password baru, validate dan hash
// 3. Update fullname dan/atau password
// 4. Return updated user info
// Returns: Updated User object atau error jika validation/db gagal.
func (u *UserUsecase) UpdateProfile(ctx context.Context, userID int, req *model.UpdateProfileRequest) (*model.User, error) {
	// 1. VALIDATE INPUT: Minimal salah satu field harus diisi
	if req.FullName == "" && req.Password == "" {
		return nil, errors.New("at least one field (full_name or password) must be provided")
	}

	// 2. FETCH USER: Cek user exists
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 3. UPDATE FULLNAME (jika diisi)
	if req.FullName != "" {
		user.FullName = req.FullName
	}

	// 4. UPDATE PASSWORD (jika diisi)
	if req.Password != "" {
		// Validate password panjang minimal 6 karakter
		if len(req.Password) < 6 {
			return nil, errors.New("password must be at least 6 characters")
		}
		// Hash password baru
		passwordHash, err := HashPassword(req.Password)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		user.PasswordHash = passwordHash
	}

	// 5. SAVE: Update ke database
	err = u.userRepo.Update(ctx, user)
	if err != nil {
		return nil, errors.New("failed to update profile")
	}

	// 6. RESPONSE: Return updated user info tanpa password hash
	return &model.User{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role:     user.Role,
		IsActive: user.IsActive,
	}, nil
}
