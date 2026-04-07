package model

// Truck merepresentasikan tabel 'trucks' di database
type Truck struct {
	ID          int    `db:"id" json:"id"`
	PlateNumber string `db:"plate_number" json:"plate_number"`
	DriverName  string `db:"driver_name" json:"driver_name"`
	IsActive    bool   `db:"is_active" json:"is_active"`
}

// UpdateTruckRequest adalah request body untuk update truck (partial update).
// Menggunakan pointers untuk optional fields sehingga bisa membedakan antara
// "field tidak dikirim" vs "field dikirim dengan value kosong/false".
type UpdateTruckRequest struct {
	PlateNumber *string `json:"plate_number"`  // nil = tidak update, empty = update ke kosong
	DriverName  *string `json:"driver_name"`   // nil = tidak update, empty = update ke kosong
	IsActive    *bool   `json:"is_active"`     // nil = tidak update, false = deactivate
}
