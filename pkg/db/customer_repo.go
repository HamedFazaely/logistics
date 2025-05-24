package db

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type CustomerRepo interface {
	CreateCustomer(ctx context.Context, cust *Customer) (*Customer, error)
	GetAllCustomers(ctx context.Context) ([]Customer, error)
}

type CustomerRepoImpl struct {
	db *sql.DB
}

func NewCustomerRepoImpl(db *sql.DB) *CustomerRepoImpl {
	return &CustomerRepoImpl{
		db: db,
	}
}

func (c *CustomerRepoImpl) CreateCustomer(ctx context.Context, cu *Customer) (*Customer, error) {
	res, err := c.db.ExecContext(ctx, "INSERT INTO customers (first_name, last_name, email) VALUES (?,?,?)", cu.FirstName, cu.LastName, cu.Email)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Customer{
		uint64(id), cu.FirstName, cu.LastName, cu.Email, cu.CreatedAt, cu.UpdatedAt,
	}, nil
}

func (c *CustomerRepoImpl) GetAllCustomers(ctx context.Context) ([]Customer, error) {
	var res []Customer
	rows, err := c.db.QueryContext(ctx,
		"SELECT id, first_name, last_name, email, created_at, updated_at FROM customers order by created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var cu Customer
		err = rows.Scan(&cu.ID, &cu.FirstName, &cu.LastName, &cu.Email, &cu.CreatedAt, &cu.UpdatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, cu)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil

}
