package usecase

import (
	"context"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
)

type OrderUsecase struct {
	orderRepo *repository.OrderRepository
	truckRepo *repository.TruckRepository
}

// NewOrderUsecase membuat instance baru dari OrderUsecase.
// Menerima dependency injection kedua repository (order dan truck) untuk validasi cross-domain.
func NewOrderUsecase(orderRepo *repository.OrderRepository, truckRepo *repository.TruckRepository) *OrderUsecase {
	return &OrderUsecase{
		orderRepo: orderRepo,
		truckRepo: truckRepo,
	}
}


// CreateOrder membuat order baru dengan validasi business rules.
// Validasi: order_number, origin, destination tidak boleh kosong.
// Business rule: status akan auto-set ke "pending" dan truck_id dimulai dari nil.
// Returns: error jika validasi gagal atau database error.
func (u *OrderUsecase) CreateOrder(ctx context.Context, o *model.Order) error {
	// Validasi: order_number wajib diisi
	if o.OrderNumber == "" {
		return ErrOrderNumberRequired
	}

	// Validasi: origin wajib diisi
	if o.Origin == "" {
		return ErrOrderOriginRequired
	}

	// Validasi: destination wajib diisi
	if o.Destination == "" {
		return ErrOrderDestinationRequired
	}

	// Business rule: status default adalah "pending"
	o.Status = "pending"

	// Simpan ke database
	return u.orderRepo.CreateOrder(ctx, o)
}

// GetByID mengambil detail single order berdasarkan ID.
// Melakukan validasi: ID harus > 0.
// Returns: pointer ke order object, atau error jika validasi/query gagal.
func (u *OrderUsecase) GetByID(ctx context.Context, id int) (*model.Order, error) {
	if id <= 0 {
		return nil, ErrOrderInvalidID
	}
	return u.orderRepo.GetByID(ctx, id)
}

// GetByOrderNumber mengambil order berdasarkan nomor order (untuk public tracking).
// Melakukan validasi: orderNumber tidak boleh kosong.
// Returns: pointer ke order object, atau error jika validasi/query gagal.
func (u *OrderUsecase) GetByOrderNumber(ctx context.Context, orderNumber string) (*model.Order, error) {
	if orderNumber == "" {
		return nil, ErrOrderNumberRequired
	}
	return u.orderRepo.GetByOrderNumber(ctx, orderNumber)
}

// AssignTruck mengassign truck ke order dengan validasi comprehensive.
// Validasi:
//   - Order dan truck existance
//   - Truck must be is_active = true
//   - Order status harus pending atau pickup
// Returns: error jika validasi gagal atau database error.
func (u *OrderUsecase) AssignTruck(ctx context.Context, orderID int, truckID int) error {
	// Validasi: order id harus valid (> 0)
	if orderID <= 0 {
		return ErrOrderInvalidID
	}

	// Validasi: truck id harus valid (> 0)
	if truckID <= 0 {
		return ErrTruckInvalidID
	}

	// Cek: truck harus ada di database
	truck, err := u.truckRepo.GetByID(ctx, truckID)
	if err != nil {
		return ErrTruckNotFound
	}

	// Cek: truck harus aktif
	if !truck.IsActive {
		return ErrTruckInactive
	}

	// Cek: order harus ada di database
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return ErrOrderNotFound
	}

	// Cek: order status harus "pending" saja untuk initial assign
	// (Setelah assign, status akan menjadi "pickup", dan tidak bisa re-assign)
	if order.Status != "pending" {
		return ErrOrderInvalidAssignStatus
	}

	// Assign truck - repository akan UPDATE status menjadi "in_transit"
	return u.orderRepo.AssignTruck(ctx, orderID, truckID)
}

// UpdateStatus mengubah status order dengan validasi state machine.
// Validasi:
//   - Order exists
//   - Status adalah salah satu dari [pending, pickup, in_transit, delivered, cancelled]
//   - Transisi status diizinkan sesuai state machine rules
// Returns: error jika validasi gagal atau database error.
func (u *OrderUsecase) UpdateStatus(ctx context.Context, id int, newStatus string) error {
	// Validasi: order id harus valid
	if id <= 0 {
		return ErrOrderInvalidID
	}

	// Validasi: status baru harus valid (salah satu dari enum values)
	validStatuses := map[string]bool{
		"pending":    true,
		"pickup":     true,
		"in_transit": true,
		"delivered":  true,
		"cancelled":  true,
	}

	if !validStatuses[newStatus] {
		return ErrOrderInvalidStatus
	}

	// Cek: order harus ada
	order, err := u.orderRepo.GetByID(ctx, id)
	if err != nil {
		return ErrOrderNotFound
	}

	// Cek: transisi status harus valid (state machine)
	if !u.isValidStatusTransition(order.Status, newStatus) {
		return ErrOrderInvalidStatusTransition
	}

	// Update status di database
	return u.orderRepo.UpdateStatus(ctx, id, newStatus)
}

// isValidStatusTransition adalah helper function yang validasi apakah transisi status diizinkan.
// Mengimplementasikan state machine logic untuk order status progression.
// Returns: true jika transisi diizinkan, false jika tidak.
func (u *OrderUsecase) isValidStatusTransition(currentStatus, newStatus string) bool {
	// Define state machine: status mana saja yang bisa transition ke status mana
	validTransitions := map[string][]string{
		"pending":    {"pickup", "cancelled"},
		"pickup":     {"in_transit", "cancelled", "pending"},
		"in_transit": {"delivered", "cancelled"},
		"delivered":  {}, // final state - tidak ada transisi keluar
		"cancelled":  {}, // final state - tidak ada transisi keluar
	}

	// Ambil list status yang diizinkan dari currentStatus
	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	// Cek apakah newStatus ada dalam list yang diizinkan
	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}

// Cancel membatalkan order dengan validasi.
// Order hanya bisa dibatalkan dari status [pending, pickup, in_transit].
// Returns: error jika validasi gagal atau database error.
func (u *OrderUsecase) Cancel(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrOrderInvalidID
	}

	// Cek: order harus ada
	order, err := u.orderRepo.GetByID(ctx, id)
	if err != nil {
		return ErrOrderNotFound
	}

	// Cek: order tidak boleh sudah delivered atau cancelled
	if order.Status == "delivered" || order.Status == "cancelled" {
		return ErrOrderCannotCancel
	}

	// Update status ke "cancelled"
	return u.orderRepo.UpdateStatus(ctx, id, "cancelled")
}
