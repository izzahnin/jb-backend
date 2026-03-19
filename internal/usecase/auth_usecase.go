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
			Role:     user.Role,
			IsActive: user.IsActive,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// Register menangani registrasi user baru (public endpoint - tidak perlu JWT).
// Flow:
// 1. Validasi input (username, password, role)
// 2. Hash password dengan bcrypt
// 3. Simpan user ke database dengan role 'customer' (user tidak bisa request admin)
// 4. Generate JWT token agar langsung bisa login
// Returns: RegisterResponse atau error jika validation/db gagal.
func (u *AuthUsecase) Register(ctx context.Context, req *model.RegisterRequest) (*model.RegisterResponse, error) {
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

	// 3. CHECK DUPLICATE: Cek username sudah ada atau belum
	existingUser, err := u.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		// Username sudah ada
		return nil, errors.New("username already exists")
	}

	// 4. HASH PASSWORD: Generate bcrypt hash
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// 5. SET ROLE: User registrasi SELALU mendapat role 'customer' (tidak bisa request admin)
	// Hanya admin yang bisa membuat user dengan role admin (via POST /admin/users)
	userRole := "customer"

	// 6. CREATE USER: Simpan user baru ke database
	newUser := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Role:         userRole,
		IsActive:     true,
	}

	err = u.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// 7. GENERATE TOKEN: Buat JWT token agar user langsung bisa login
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &middleware.Claims{
		UserID:   newUser.ID,
		Username: newUser.Username,
		Role:     newUser.Role,
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
	return &model.RegisterResponse{
		Token:     tokenString,
		ExpiresAt: expirationTime.Unix(),
		User: model.User{
			ID:       newUser.ID,
			Username: newUser.Username,
			Role:     newUser.Role,
		},
	}, nil
}

// CreateUser menangani pembuatan user baru oleh admin (admin-protected endpoint).
// Flow:
// 1. Validasi input (username, password, role)
// 2. Validasi role adalah 'admin' atau 'customer'
// 3. Hash password dengan bcrypt
// 4. Simpan user ke database
// 5. Return user info (tidak return token - admin yang beri password ke user)
// Returns: User model atau error jika validation/db gagal.
func (u *AuthUsecase) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
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

	// 6. CREATE USER: Simpan user baru ke database dengan role yang dipilih admin
	newUser := &model.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Role:         req.Role,
		IsActive:     req.IsActive,
	}

	err = u.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	// 7. RESPONSE: Return user info tanpa password hash
	// Client (admin) akan memberi username+password ke user via secure channel
	return &model.User{
		ID:       newUser.ID,
		Username: newUser.Username,
		Role:     newUser.Role,
		IsActive: newUser.IsActive,
	}, nil
}

// HashPassword mengkonversi plain text password menjadi bcrypt hash.
// Digunakan saat create user (tidak dipakai di login, tapi bisa dipakai di registration).
// Cost parameter menentukan complexity (10-12 recommended untuk production).
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hash), err
}
