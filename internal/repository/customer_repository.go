package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/izzahnin/jalur-berlian-backend/internal/model"
	"github.com/jmoiron/sqlx"
)

// CustomerRepository provides CRUD operations for customers table.
type CustomerRepository struct {
	db *sqlx.DB
}

func NewCustomerRepository(db *sqlx.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, c *model.Customer) error {
	query := `INSERT INTO customers (company_name, pic_name, phone, email, address, npwp, is_active)
	          VALUES ($1, $2, $3, $4, $5, $6, $7)
	          RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		c.CompanyName, c.PICName, c.Phone, c.Email, c.Address, c.NPWP, true,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *CustomerRepository) FetchAll(ctx context.Context) ([]model.Customer, error) {
	var customers []model.Customer
	query := `SELECT id, company_name, pic_name, phone, email, address, npwp, is_active, created_at
	          FROM customers
	          WHERE is_active = true
	          ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &customers, query); err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, id int) (*model.Customer, error) {
	var c model.Customer
	query := `SELECT id, company_name, pic_name, phone, email, address, npwp, is_active, created_at
	          FROM customers
	          WHERE id = $1 AND is_active = true`
	if err := r.db.GetContext(ctx, &c, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("customer not found")
		}
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) Update(ctx context.Context, id int, c *model.Customer) error {
	query := `UPDATE customers
	          SET company_name = $1,
	              pic_name = $2,
	              phone = $3,
	              email = $4,
	              address = $5,
	              npwp = $6,
	              is_active = $7
	          WHERE id = $8 AND is_active = true`
	_, err := r.db.ExecContext(ctx, query,
		c.CompanyName, c.PICName, c.Phone, c.Email, c.Address, c.NPWP, c.IsActive, id,
	)
	return err
}

func (r *CustomerRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE customers SET is_active = false WHERE id = $1 AND is_active = true`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
