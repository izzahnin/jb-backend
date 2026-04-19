package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/izzahnin/jalur-berlian-backend/internal/middleware"
	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthUsecase menghandle business logic untuk authentication.
// Memanggil repository untuk database access dan signature token dengan JWT.
type AuthUsecase struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

// NewAuthUsecase membuat instance baru dari AuthUsecase.
// jwtSecret adalah secret key untuk generate/verify JWT token (keep it secret!).
func NewAuthUsecase(userRepo *repository.UserRepository, jwtSecret string) *AuthUsecase {
	return &AuthUsecase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Login memvalidasi username/password dan mengembalikan JWT token jika valid.
// Flow:
// 1. Cari user di database berdasarkan username
// 2. Bandingkan password plain text dengan bcrypt hash
// 3. Jika cocok, generate JWT token dengan user info di claims
// 4. Return token dan expiration time
// Returns: LoginResponse atau error jika validation gagal.
func (u *AuthUsecase) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	// 1. VALIDATE INPUT: Username dan password wajib diisi
	if req.Username == "" {
		return nil, errors.New("username required")
	}
	if req.Password == "" {
		return nil, errors.New("password required")
	}

	// 2. LOOKUP: Cari user di database berdasarkan username
	user, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// User tidak ditemukan → return generic error agar tidak expose user existence
		return nil, errors.New("invalid username or password")
	}

	// 3. CHECK ACTIVE: Pastikan user tidak di-deactivate
	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	// 4. VERIFY PASSWORD: Compare plain text password dengan bcrypt hash
	// bcrypt.CompareHashAndPassword return nil jika cocok, error jika tidak
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		// Password salah → return generic error agar tidak expose valid username
		return nil, errors.New("invalid username or password")
	}

	// 5. GENERATE TOKEN: Buat JWT token dengan user info
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid 24 jam
	claims := &middleware.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 6. SIGN: Generate JWT token dengan HMAC-SHA256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// 7. RESPONSE: Return token dan expiration
	return &model.LoginResponse{
		Token:     tokenString,
		ExpiresAt: expirationTime.Unix(),
		User: model.User{
			ID:       user.ID,
			Username: user.Username,
			FullName: user.FullName,
			Role:     user.Role,
			IsActive: user.IsActive,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// AdminSetup menangani pembuatan super_admin pertama kali (one-time setup).
// Endpoint PUBLIK - tidak perlu authentication, hanya bisa dijalankan SEKALI.
// Setelah ada super_admin, gunakan POST /admin/users (dengan JWT super_admin) untuk create user lainnya.
//
// Workflow Admin Setup:
// 1. POST /admin/setup (one-time, public) → create super_admin + return JWT
// 2. POST /auth/login (public) → login dengan super_admin credentials
// 3. POST /admin/users (protected, super_admin only) → create admin_sales atau admin_ops users
//
// Flow Setup:
// 1. Cek apakah sudah ada super_admin user di sistem
// 2. Jika sudah ada, return error (endpoint sudah disabled secara logic)
// 3. Jika belum ada, validasi input (username, password)
// 4. Hash password dengan bcrypt
// 5. Simpan super_admin user pertama
// 6. Generate JWT token dan return (langsung bisa login tanpa POST /auth/login)
// Returns: LoginResponse (token + user info) atau error jika super_admin sudah ada atau validation gagal.
func (u *AuthUsecase) AdminSetup(ctx context.Context, req *model.AdminSetupRequest) (*model.LoginResponse, error) {
	// 1. CHECK EXISTENCE: Cek apakah sudah ada admin user
	adminExists, err := u.userRepo.AdminExists(ctx)
	if err != nil {
		return nil, errors.New("database error while checking admin existence")
	}
	if adminExists {
		// Admin sudah ada - endpoint tidak bisa dijalankan lagi
		return nil, errors.New("admin user already exists")
	}

	// 2. VALIDATE: Username wajib diisi dan non-empty
	if req.Username == "" {
		return nil, errors.New("username required")
	}

	// 3. VALIDATE: Password wajib diisi dan minimal 6 karakter
	if req.Password == "" {
		return nil, errors.New("password required")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// 4. CHECK DUPLICATE: Cek username sudah ada atau belum
	existingUser, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 5. HASH PASSWORD: Generate bcrypt hash
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 6. SET DEFAULT FULLNAME: Jika fullname kosong, gunakan username sebagai display name
	fullName := req.FullName
	if fullName == "" {
		fullName = req.Username
	}

	// 7. CREATE ADMIN USER WITH ROLE 'super_admin'
	newAdmin := &model.User{
		Username:     req.Username,
		FullName:     fullName,
		PasswordHash: passwordHash,
		Role:         "super_admin",
		IsActive:     true,
	}

	err = u.userRepo.Create(ctx, newAdmin)
	if err != nil {
		return nil, errors.New("failed to create admin user")
	}

	// 7. GENERATE TOKEN: Buat JWT token agar admin langsung bisa login
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &middleware.Claims{
		UserID:   newAdmin.ID,
		Username: newAdmin.Username,
		Role:     newAdmin.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	// 8. RESPONSE: Return token dan user info
	return &model.LoginResponse{
		Token:     tokenString,
		ExpiresAt: expirationTime.Unix(),
		User: model.User{
			ID:       newAdmin.ID,
			Username: newAdmin.Username,
			FullName: newAdmin.FullName,
			Role:     newAdmin.Role,
			IsActive: newAdmin.IsActive,
		},
	}, nil
}

// HashPassword mengkonversi plain text password menjadi bcrypt hash.
// Digunakan saat create user (tidak dipakai di login, tapi bisa dipakai di registration).
// Cost parameter menentukan complexity (10-12 recommended untuk production).
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hash), err
}
