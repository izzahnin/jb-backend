package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Driver PostgreSQL
)

// NewPostgres membuat pool koneksi database menggunakan DSN (Data Source Name).
// Koneksi aktual diverifikasi terpisah saat startup agar bisa memakai context dan retry.
func NewPostgres(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Mengatur pool koneksi (Standar industri agar DB tidak overload)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return db, nil
}
