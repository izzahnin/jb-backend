package model

import "time"

// Location menyimpan titik koordinat untuk sebuah truk
type Location struct {
	ID        int       `db:"id" json:"-"`
	TruckID   int       `db:"truck_id" json:"truck_id"`
	Latitude  float64   `db:"latitude" json:"latitude"`
	Longitude float64   `db:"longitude" json:"longitude"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
