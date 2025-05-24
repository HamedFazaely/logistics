package db

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"gitlab.com/Hamed1984/logistics/pkg/conf"
)

type OrderRepo interface {
	CreateOrder(ctx context.Context, ord *Order) (*Order, error)
	GetOrderByID(ctx context.Context, id uint64) (*Order, error)
	UpdateOrderStatus(ctx context.Context, id uint64, stat string) error
	SetPickupSMSSent(ctx context.Context, id uint64) error
	SetOrderPickedUp(ctx context.Context, id uint64) (*OrderWithProvider, error)
}

type OrderRepoImpl struct {
	db     *sql.DB
	config *conf.Configuration
}

func NewOrderRepoImpl(db *sql.DB, conf *conf.Configuration) *OrderRepoImpl {
	return &OrderRepoImpl{
		db:     db,
		config: conf,
	}
}

func (o *OrderRepoImpl) CreateOrder(ctx context.Context, ord *Order) (*Order, error) {
	res, err := o.db.ExecContext(ctx, "INSERT INTO orders (customer_id, s_phone, r_phone, status, address, provider_id) VALUES (?,?,?,?,?,?)",
		ord.CutomerID, ord.SenderPhone, ord.ReceiverPhone, Pending, ord.Address, ord.ProviderID)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Order{
		ID:            uint64(id),
		CutomerID:     ord.CutomerID,
		SenderPhone:   ord.SenderPhone,
		ReceiverPhone: ord.ReceiverPhone,
		Status:        Pending,
		Address:       ord.Address,
		ProviderID:    ord.ProviderID,
	}, nil
}

func (o *OrderRepoImpl) GetOrderByID(ctx context.Context, id uint64) (*Order, error) {
	var ord Order
	row := o.db.QueryRowContext(ctx, "SELECT * FROM orders WHERE id = ? ", id)
	err := row.Scan(&ord.ID, &ord.CutomerID, &ord.SenderPhone, &ord.ReceiverPhone, &ord.Status, &ord.PickupSMSSent, &ord.PickupTime, &ord.DeliveryTime, &ord.Address, &ord.ProviderID, &ord.CreatedAt, &ord.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ord, nil
}

func (o *OrderRepoImpl) UpdateOrderStatus(ctx context.Context, id uint64, stat string) error {
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer o.rollBack(tx)
	ord, err := o.getOrderByIDTxn(ctx, tx, id)
	if err != nil {
		return err
	}
	if ord.Status == PickedUp && stat == ProviderSeen {
		return ErrOrderAlreadyPickedUp
	}
	if ord.Status == Delivered {
		return ErrOrderAlreadyDelivered
	}

	if stat == Pending {
		return ErrOrderAlreadyPickedUp
	}

	if stat == Delivered {
		res, err := tx.ExecContext(ctx, "UPDATE orders SET status = ?, delivery_time = ? WHERE id = ? AND updated_at = ?", stat, time.Now(), id, ord.UpdatedAt)
		if err != nil {
			return err
		}
		ra, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if ra == 0 {
			return ErrOrderAlreadyDelivered
		}
		return tx.Commit()
	}

	res, err := tx.ExecContext(ctx, "UPDATE orders SET status = ? WHERE id = ? AND updated_at = ?", stat, id, ord.UpdatedAt)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrOrderStatusAlreadySet
	}
	return tx.Commit()

}

func (o *OrderRepoImpl) SetPickupSMSSent(ctx context.Context, id uint64) error {
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer o.rollBack(tx)
	ord, err := o.getOrderByIDTxn(ctx, tx, id)
	if err != nil {
		return err
	}
	if ord.PickupSMSSent {
		return ErrPickupSMSAlreadySent
	}
	res, err := tx.ExecContext(ctx, "UPDATE orders SET pickup_sms_sent = ? WHERE id = ? AND updated_at = ?", true, id, ord.UpdatedAt)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrPickupSMSAlreadySent
	}
	return tx.Commit()

}

func (o *OrderRepoImpl) SetOrderPickedUp(ctx context.Context, id uint64) (*OrderWithProvider, error) {
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer o.rollBack(tx)

	ord, err := o.getOrderByIDTxn(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if ord.Status == PickedUp {
		return nil, ErrOrderAlreadyPickedUp
	}

	if ord.Status == Delivered {
		return nil, ErrOrderAlreadyDelivered
	}

	if ord.CreatedAt.Add(time.Duration(o.config.OrderExpirationHours) * time.Hour).Before(time.Now()) {
		return nil, ErrOrderExpired
	}

	res, err := tx.ExecContext(ctx, "UPDATE orders SET status = ?, pickup_time = ? WHERE id = ? AND updated_at = ?", PickedUp, time.Now(), id, ord.UpdatedAt)
	if err != nil {
		return nil, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if ra == 0 {
		return nil, ErrOrderAlreadyPickedUp
	}

	var ordWP OrderWithProvider
	row := tx.QueryRowContext(context.Background(), "SELECT orders.id, orders.pickup_sms_sent, orders.status,  providers.status_endpoint, orders.r_phone, providers.pickup_endpoint FROM providers INNER JOIN orders ON providers.id = orders.provider_id WHERE orders.id = ?", id)
	err = row.Scan(&ordWP.OrderID, &ordWP.PickupSMSSent, &ordWP.Status, &ordWP.StatusEP, &ordWP.ReceiverPhone, &ordWP.NotifyEP)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return &ordWP, nil
}

func (o *OrderRepoImpl) getOrderByIDTxn(ctx context.Context, tx *sql.Tx, id uint64) (*Order, error) {
	var ord Order
	row := tx.QueryRowContext(ctx, "SELECT * FROM orders WHERE id = ? ", id)
	err := row.Scan(&ord.ID, &ord.CutomerID, &ord.SenderPhone, &ord.ReceiverPhone, &ord.Status, &ord.PickupSMSSent, &ord.PickupTime, &ord.DeliveryTime, &ord.Address, &ord.ProviderID, &ord.CreatedAt, &ord.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &ord, nil
}

func (o *OrderRepoImpl) rollBack(tx *sql.Tx) {
	err := tx.Rollback()
	if err != nil {
		if !errors.Is(err, sql.ErrTxDone) {
			log.Printf("an error ocuured during transaxtion rollback: %s\n", err.Error())
		}
	}

}
