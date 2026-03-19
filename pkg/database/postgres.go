package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Driver PostgreSQL
)

// NewPostgres membuka koneksi ke database menggunakan DSN (Data Source Name)
func NewPostgres(dsn string) (*sqlx.DB, error) {
	// sqlx.Connect langsung melakukan Open dan Ping sekaligus
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Mengatur pool koneksi (Standar industri agar DB tidak overload)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return db, nil
}
