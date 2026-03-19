package model

import "time"

type Order struct {
	ID          int       `db:"id" json:"id"`
	OrderNumber string    `db:"order_number" json:"order_number"`
	TruckID     *int      `db:"truck_id" json:"truck_id"` // Menggunakan pointer agar bisa NULL
	Origin      string    `db:"origin" json:"origin"`
	Destination string    `db:"destination" json:"destination"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
