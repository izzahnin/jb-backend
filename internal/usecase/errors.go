package usecase

import "errors"

// Standard error definitions untuk seluruh usecase layer
// Menggunakan error variables agar dapat di-compare dengan == dalam error handling

var (
	// Truck-related errors
	ErrTruckInvalidID     = errors.New("truck id tidak valid")
	ErrTruckNotFound      = errors.New("truck tidak ditemukan")
	ErrTruckInactive      = errors.New("truck tidak aktif")
	ErrTruckRequired      = errors.New("truck wajib diisikan")
	ErrTruckPlateRequired = errors.New("nomor plat truck wajib diisi")
	ErrTruckTypeRequired  = errors.New("tipe truck wajib diisi")
	ErrTruckHasActiveTrips = errors.New("truck masih memiliki trip aktif, tidak bisa dinonaktifkan")
	ErrTruckOnActiveTrip   = errors.New("truck sedang dalam trip aktif, status tidak bisa diubah")

	// Driver-related errors
	ErrDriverInvalidID = errors.New("driver id tidak valid")
	ErrDriverNotFound  = errors.New("driver tidak ditemukan")
	ErrDriverInactive  = errors.New("driver tidak aktif")
	ErrDriverNameRequired    = errors.New("nama driver wajib diisi")
	ErrDriverLicenseRequired = errors.New("nomor SIM wajib diisi")
	ErrDriverPhoneRequired   = errors.New("nomor telepon driver wajib diisi")
	ErrDriverOnActiveTrip    = errors.New("driver sedang dalam trip aktif, status tidak bisa diubah")

	// Customer-related errors
	ErrCustomerInvalidID = errors.New("customer id tidak valid")
	ErrCustomerNotFound  = errors.New("customer tidak ditemukan")
	ErrCustomerCompanyRequired = errors.New("nama perusahaan wajib diisi")
	ErrCustomerPICRequired     = errors.New("nama PIC wajib diisi")
	ErrCustomerPhoneRequired   = errors.New("nomor telepon wajib diisi")

	// Order-related errors
	ErrOrderInvalidID              = errors.New("order id tidak valid")
	ErrOrderNotFound               = errors.New("order tidak ditemukan")
	ErrOrderNumberRequired         = errors.New("nomor order wajib diisi")
	ErrOrderCustomerRequired       = errors.New("customer wajib diisi")
	ErrOrderOriginRequired         = errors.New("lokasi asal wajib diisi")
	ErrOrderDestinationRequired    = errors.New("lokasi tujuan wajib diisi")
	ErrOrderTotalContainersInvalid = errors.New("total kontainer harus 1")
	ErrOrderInvalidStatus          = errors.New("status order tidak valid")
	ErrOrderInvalidStatusTransition = errors.New("transisi status tidak diizinkan")
	ErrOrderCannotCancel           = errors.New("order tidak bisa dibatalkan dari status ini")

	// Trip-related errors
	ErrTripInvalidID               = errors.New("trip id tidak valid")
	ErrTripNotFound                = errors.New("trip tidak ditemukan")
	ErrTripAlreadyExistsForOrder   = errors.New("order ini sudah memiliki trip")
	ErrTripNumberRequired          = errors.New("nomor surat jalan wajib diisi")
	ErrTripInvalidStatus           = errors.New("status trip tidak valid")
	ErrTripInvalidStatusTransition = errors.New("transisi status trip tidak diizinkan")
	ErrTripContainerRequired       = errors.New("nomor kontainer wajib diisi")
	ErrTripSealRequired            = errors.New("nomor segel wajib diisi")

	// Location-related errors
	ErrLocationInvalidTripID    = errors.New("trip id untuk location tidak valid")
	ErrLocationInvalidCoords    = errors.New("koordinat latitude/longitude tidak valid")
	ErrLocationTripNotInTransit = errors.New("lokasi hanya bisa dikirim saat trip sedang in_transit")
	ErrLocationTripDelivered    = errors.New("lokasi tidak bisa dikirim untuk trip yang sudah delivered")
	ErrLocationThrottled        = errors.New("interval terlalu singkat, tunggu 30 detik sebelum kirim lokasi lagi")

	// Rate limiting errors
	ErrRateLimitExceeded = errors.New("terlalu banyak percobaan, coba lagi nanti")

	// Validation errors
	ErrValidationFailed = errors.New("validasi data gagal")
)
