package usecase

import (
	"context"
	"time"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/izzahnin/jalur-berlian-backend/internal/repository"
)

type OrderUsecase struct {
	orderRepo *repository.OrderRepository
	customerRepo *repository.CustomerRepository
}

// NewOrderUsecase membuat instance baru dari OrderUsecase.
// Menerima dependency injection kedua repository (order dan truck) untuk validasi cross-domain.
func NewOrderUsecase(orderRepo *repository.OrderRepository, customerRepo *repository.CustomerRepository) *OrderUsecase {
	return &OrderUsecase{
		orderRepo: orderRepo,
		customerRepo: customerRepo,
	}
}


// CreateOrder membuat order baru dengan validasi business rules.
// Validasi: origin, destination, customer, dan total containers tidak boleh kosong/invalid.
// Business rule: order_number dibuat otomatis berurutan dan status akan auto-set ke "pending".
// Returns: error jika validasi gagal atau database error.
func (u *OrderUsecase) CreateOrder(ctx context.Context, o *model.Order) error {
	// Validasi: origin wajib diisi
	if o.Origin == "" {
		return ErrOrderOriginRequired
	}

	// Validasi: destination wajib diisi
	if o.Destination == "" {
		return ErrOrderDestinationRequired
	}

	if o.CustomerID <= 0 {
		return ErrOrderCustomerRequired
	}

	if o.TotalContainers != 1 {
		return ErrOrderTotalContainersInvalid
	}

	if _, err := u.customerRepo.GetByID(ctx, o.CustomerID); err != nil {
		return ErrCustomerNotFound
	}

	o.Status = "pending"
	o.IsActive = true
	o.OrderDate = time.Now().UTC()

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
		"pending":   true,
		"partial":   true,
		"completed": true,
		"cancelled": true,
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
		"pending":   {"partial", "cancelled"},
		"partial":   {"completed", "cancelled"},
		"completed": {},
		"cancelled": {},
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
	if order.Status == "completed" || order.Status == "cancelled" {
		return ErrOrderCannotCancel
	}

	if err := u.orderRepo.UpdateStatus(ctx, id, "cancelled"); err != nil {
		return err
	}

	// Soft delete agar tidak tampil di list aktif, tapi data tetap bisa dipulihkan.
	return u.orderRepo.Delete(ctx, id)
}
