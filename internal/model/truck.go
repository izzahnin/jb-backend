package model

// Truck merepresentasikan tabel 'trucks' di database
type Truck struct {
	ID          int    `db:"id" json:"id"`
	PlateNumber string `db:"plate_number" json:"plate_number"`
	DriverName  string `db:"driver_name" json:"driver_name"`
	IsActive    bool   `db:"is_active" json:"is_active"`
}
