package model

import "time"

// Location menyimpan titik koordinat untuk sebuah truk
type Location struct {
	ID        int64     `db:"id" json:"id"`
	TripID    *int      `db:"trip_id" json:"trip_id" example:"1"`
	Latitude  float64   `db:"latitude" json:"latitude" example:"-6.200000"`
	Longitude float64   `db:"longitude" json:"longitude" example:"106.816666"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CreateLocationRequest is the request DTO for posting trip location updates.
type CreateLocationRequest struct {
	Lat   float64  `json:"lat" example:"-6.200000" binding:"required"`
	Lon   float64  `json:"lon" example:"106.816666" binding:"required"`
	Ts    string   `json:"ts" example:"2026-04-19T08:30:00Z"`
}
