package usecase

import (
	"context"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
)

type TruckUsecase struct {
	repo *repository.TruckRepository
}

// NewTruckUsecase membuat instance baru dari TruckUsecase.
// Menerima dependency injection repository untuk akses data truck.
func NewTruckUsecase(repo *repository.TruckRepository) *TruckUsecase {
	return &TruckUsecase{repo: repo}
}

// RegisterTruck menambahkan truck baru ke sistem dengan validasi.
// Validasi mencakup: plate_number dan driver_name tidak boleh kosong.
// Default: truck akan di-set sebagai is_active = true.
// Returns: error jika validasi gagal atau database error.
func (u *TruckUsecase) RegisterTruck(ctx context.Context, t *model.Truck) error {
	// Validasi: plate_number wajib diisi
	if t.PlateNumber == "" {
		return ErrTruckPlateRequired
	}

	// Validasi: driver_name wajib diisi
	if t.DriverName == "" {
		return ErrTruckDriverRequired
	}

	// Business rule: set truck baru sebagai aktif
	t.IsActive = true

	return u.repo.Create(ctx, t)
}

// GetByID mengambil detail single truck berdasarkan ID.
// Melakukan validasi: ID harus > 0.
// Returns: pointer ke truck object, atau error jika validasi/query gagal.
func (u *TruckUsecase) GetByID(ctx context.Context, id int) (*model.Truck, error) {
	if id <= 0 {
		return nil, ErrTruckInvalidID
	}
	return u.repo.GetByID(ctx, id)
}

// Update mengubah informasi truck (driver_name, is_active) berdasarkan ID.
// Melakukan validasi: ID harus > 0, driver_name tidak boleh kosong.
// Returns: error jika validasi gagal atau database error.
func (u *TruckUsecase) Update(ctx context.Context, id int, t *model.Truck) error {
	// Validasi: truck id harus valid
	if id <= 0 {
		return ErrTruckInvalidID
	}

	// Validasi: driver_name tidak boleh kosong
	if t.DriverName == "" {
		return ErrTruckDriverRequired
	}

	return u.repo.Update(ctx, id, t)
}

// Deactivate melakukan soft delete truck dengan mengeset is_active = false.
// Melakukan validasi: ID harus > 0.
// Catatan: Dapat ditambahkan validasi tsb. truck tidak sedang mengerjakan order aktif.
// Returns: error jika validasi gagal atau database error.
func (u *TruckUsecase) Deactivate(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrTruckInvalidID
	}
	return u.repo.Delete(ctx, id)
}
