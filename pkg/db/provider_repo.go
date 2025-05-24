package db

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type ProviderRepo interface {
	CreateProvider(ctx context.Context, prov *Provider) (*Provider, error)
	GetAllProviders(ctx context.Context) ([]Provider, error)
	GetProviderByID(ctx context.Context, id uint64) (*Provider, error)
	GetAverageDelivery(ctx context.Context) ([]DeliveryAvg, error)
}

type ProviderRepoImpl struct {
	db *sql.DB
}

func NewProverRepoImpl(db *sql.DB) *ProviderRepoImpl {
	return &ProviderRepoImpl{
		db: db,
	}
}

func (p *ProviderRepoImpl) GetProviderByID(ctx context.Context, id uint64) (*Provider, error) {
	row := p.db.QueryRowContext(ctx, "SELECT * FROM providers WHERE id = ?", id)
	var res Provider
	err := row.Scan(&res.ID, &res.Name, &res.StatusEP, &res.PickupEP, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (p *ProviderRepoImpl) CreateProvider(ctx context.Context, prov *Provider) (*Provider, error) {
	res, err := p.db.ExecContext(ctx,
		"INSERT INTO providers (name, status_endpoint, pickup_endpoint) VALUES (?,?,?)",
		prov.Name, prov.StatusEP, prov.PickupEP)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Provider{
		ID:       uint64(id),
		Name:     prov.Name,
		StatusEP: prov.StatusEP,
		PickupEP: prov.PickupEP,
	}, nil
}

func (p *ProviderRepoImpl) GetAllProviders(ctx context.Context) ([]Provider, error) {
	var res []Provider
	rows, err := p.db.QueryContext(ctx, "SELECT id, name, status_endpoint, pickup_endpoint, created_at, updated_at FROM providers ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var prov Provider
		err := rows.Scan(&prov.ID, &prov.Name, &prov.StatusEP, &prov.PickupEP, &prov.CreatedAt, &prov.UpdatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, prov)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}

func (p *ProviderRepoImpl) GetAverageDelivery(ctx context.Context) ([]DeliveryAvg, error) {
	query := "SELECT AVG(TIMESTAMPDIFF(MINUTE,orders.pickup_time,orders.delivery_time)) as avg_del , providers.name FROM providers INNER JOIN orders on providers.id = orders.provider_id WHERE orders.pickup_time IS NOT NULL AND orders.delivery_time IS NOT NULL AND DATE(orders.created_at) > TIMESTAMPADD(WEEK,-1,DATE(orders.created_at)) GROUP BY providers.name ORDER BY avg_del DESC"

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []DeliveryAvg
	for rows.Next() {
		var d DeliveryAvg
		err := rows.Scan(&d.Avg, &d.ProviderName)
		if err != nil {
			return nil, err
		}
		res = append(res, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
