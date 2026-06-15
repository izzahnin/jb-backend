package usecase

import (
	"context"
	"strings"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
)

type TruckUsecase struct {
	repo     *repository.TruckRepository
	tripRepo *repository.TripRepository
}

// NewTruckUsecase membuat instance baru dari TruckUsecase.
// Menerima dependency injection repository untuk akses data truck.
func NewTruckUsecase(repo *repository.TruckRepository, tripRepo *repository.TripRepository) *TruckUsecase {
	return &TruckUsecase{
		repo:     repo,
		tripRepo: tripRepo,
	}
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

	if t.TruckType == "" {
		return ErrTruckTypeRequired
	}

	// Business rule: set truck baru sebagai aktif
	t.IsActive = true
	if t.Status == "" {
		t.Status = "available"
	}
	t.Status = normalizeTruckStatus(t.Status)

	validStatus := map[string]bool{"available": true, "on_duty": true, "maintenance": true}
	if !validStatus[t.Status] {
		return ErrValidationFailed
	}

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
	if req.PlateNumber != nil {
		existing.PlateNumber = *req.PlateNumber
	}

	if req.TruckType != nil {
		existing.TruckType = *req.TruckType
	}

	if req.Status != nil {
		existing.Status = normalizeTruckStatus(*req.Status)
	}

	// is_active: update hanya jika pointer diisi
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	// Final validation
	if existing.PlateNumber == "" {
		return ErrTruckPlateRequired
	}
	if existing.TruckType == "" {
		return ErrTruckTypeRequired
	}

	validStatus := map[string]bool{"available": true, "on_duty": true, "maintenance": true}
	if !validStatus[existing.Status] {
		return ErrValidationFailed
	}

	return u.repo.Update(ctx, id, existing)
}

// Deactivate melakukan soft delete truck dengan mengeset is_active = false.
// Melakukan validasi: ID harus > 0 dan truck tidak sedang mengerjakan trip aktif.
// Jika truck masih punya trip aktif (status: pickup atau in_transit), deactivate akan ditolak.
// Returns: error jika validasi gagal atau database error.
func (u *TruckUsecase) Deactivate(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrTruckInvalidID
	}

	// Cek apakah truck masih punya trip aktif
	activeTripsCount, err := u.tripRepo.CountActiveByTruckID(ctx, id)
	if err != nil {
		return err
	}
	if activeTripsCount > 0 {
		return ErrTruckHasActiveTrips
	}

	return u.repo.Delete(ctx, id)
}

func normalizeTruckStatus(status string) string {
	trimmed := strings.TrimSpace(strings.ToLower(status))
	if trimmed == "in_use" {
		return "on_duty"
	}
	return trimmed
}
