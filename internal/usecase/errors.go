package usecase

import "errors"

// Standard error definitions untuk seluruh usecase layer
// Menggunakan error variables agar dapat di-compare dengan == dalam error handling

var (
	// Truck-related errors
	ErrTruckInvalidID   = errors.New("truck id tidak valid")
	ErrTruckNotFound    = errors.New("truck tidak ditemukan")
	ErrTruckInactive    = errors.New("truck tidak aktif, tidak bisa ditugaskan ke order")
	ErrTruckRequired    = errors.New("truck wajib diisikan")
	ErrTruckPlateRequired = errors.New("nomor plat truck wajib diisi")
	ErrTruckDriverRequired = errors.New("nama driver truck wajib diisi")

	// Order-related errors
	ErrOrderInvalidID              = errors.New("order id tidak valid")
	ErrOrderNotFound               = errors.New("order tidak ditemukan")
	ErrOrderNumberRequired         = errors.New("nomor order wajib diisi")
	ErrOrderOriginRequired         = errors.New("lokasi asal wajib diisi")
	ErrOrderDestinationRequired    = errors.New("lokasi tujuan wajib diisi")
	ErrOrderInvalidStatus          = errors.New("status order tidak valid")
	ErrOrderInvalidStatusTransition = errors.New("transisi status tidak diizinkan")
	ErrOrderCannotCancel           = errors.New("order tidak bisa dibatalkan dari status ini")
	ErrOrderInvalidAssignStatus    = errors.New("order harus dalam status pending atau pickup untuk bisa ditugaskan truk")

	// Location-related errors
	ErrLocationInvalidTruckID = errors.New("truck id untuk location tidak valid")
	ErrLocationInvalidCoords  = errors.New("koordinat latitude/longitude tidak valid")

	// Validation errors
	ErrValidationFailed = errors.New("validasi data gagal")
)
