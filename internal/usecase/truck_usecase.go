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

// Update mengubah informasi truck (plate_number, driver_name, is_active) berdasarkan ID.
// Mendukung partial updates - hanya field yang disediakan yang akan diupdate.
// Validasi: ID harus > 0. Setelah merge, plate_number dan driver_name tidak boleh kosong.
// Returns: error jika validasi gagal atau database error.
func (u *TruckUsecase) Update(ctx context.Context, id int, req *model.UpdateTruckRequest) error {
	// Validasi: truck id harus valid
	if id <= 0 {
		return ErrTruckInvalidID
	}

	// Fetch current truck data
	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Merge updates: hanya update field yang diisi (pointer tidak nil)
	// plate_number: update hanya jika pointer diisi
	if req.PlateNumber != nil {
		existing.PlateNumber = *req.PlateNumber
	}

	// driver_name: update hanya jika pointer diisi
	if req.DriverName != nil {
		existing.DriverName = *req.DriverName
	}

	// is_active: update hanya jika pointer diisi
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	// Final validation: pastikan setelah merge, kedua field required tidak kosong
	if existing.PlateNumber == "" {
		return ErrTruckPlateRequired
	}
	if existing.DriverName == "" {
		return ErrTruckDriverRequired
	}

	return u.repo.Update(ctx, id, existing)
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
