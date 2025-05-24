package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.com/Hamed1984/logistics/pkg/conf"
	"gitlab.com/Hamed1984/logistics/pkg/db"
	"gitlab.com/Hamed1984/logistics/pkg/logging"
	"gitlab.com/Hamed1984/logistics/pkg/mq"
	"gitlab.com/Hamed1984/logistics/pkg/providerclient"
	"go.uber.org/zap"
)

type Context struct {
	config       *conf.Configuration
	customerRepo db.CustomerRepo
	proverRepo   db.ProviderRepo
	ordersRepo   db.OrderRepo
	smsMq        mq.SMSPublisherConsumer
	statusMQ     mq.StatusPublisherReceiver
	cacnel       <-chan struct{}
	smsCh        chan *mq.SMSMsg
	smsErrCh     chan *mq.SMSPubError
	statusCh     chan *mq.StatusMsg
	statusErrCh  chan *mq.StatusPubError
}

func NewContext(c *conf.Configuration, cancel <-chan struct{}) (*Context, error) {
	conn, err := db.NewMysqlConnection(c)
	if err != nil {
		return nil, err
	}
	customerRepo := db.NewCustomerRepoImpl(conn)
	providerRepo := db.NewProverRepoImpl(conn)
	orderRepo := db.NewOrderRepoImpl(conn, c)

	smsMQ, err := mq.NewSMS(c, cancel)
	if err != nil {
		return nil, err
	}

	smsCh := make(chan *mq.SMSMsg)
	smsErrCh := make(chan *mq.SMSPubError)

	go func() {

		for x := range smsErrCh {
			logging.GetLogger(c).Error(x.Err.Error())
			logging.GetLogger(c).Info("retry publishing message", zap.Uint64("order_id", x.Msg.OrderID))
			smsCh <- x.Msg
		}

	}()

	smsMQ.StartPublisher(smsCh, smsErrCh)
	smsMQ.StartConsume(func(d amqp.Delivery, done <-chan struct{}) error {
		select {
		case <-done:
			return nil
		default:
			var msg mq.SMSMsg
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				logging.GetLogger(c).Error(err.Error())
				d.Nack(false, false)
				return nil
			}
			ctx, canceF := context.WithTimeout(context.Background(), 10*time.Second)
			defer canceF()
			err = orderRepo.SetPickupSMSSent(ctx, msg.OrderID)
			if err != nil {
				if err == db.ErrPickupSMSAlreadySent {
					d.Ack(false)
					return nil
				}
				logging.GetLogger(c).Error(err.Error())
				d.Nack(false, true)
				return nil
			}
			d.Ack(false)

		}
		return nil

	})

	statMQ, err := mq.NewStatusMQ(c, cancel)
	if err != nil {
		return nil, err
	}

	statCh := make(chan *mq.StatusMsg)
	statErrCh := make(chan *mq.StatusPubError)

	go func() {

		for x := range statErrCh {
			logging.GetLogger(c).Error(x.Err.Error())
			logging.GetLogger(c).Info("retry publising message", zap.Uint64("order_id", x.Msg.OrderID))
			statCh <- x.Msg
		}

	}()

	statMQ.StartPublisher(statCh, statErrCh)
	statMQ.StartConsume(func(d amqp.Delivery, done <-chan struct{}) error {
		select {
		case <-done:
			return nil
		default:
			var msg mq.StatusMsg
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				logging.GetLogger(c).Error(err.Error())
				d.Ack(false)
				return nil
			}
			ctx, cancelF := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancelF()
			err = orderRepo.UpdateOrderStatus(ctx, msg.OrderID, msg.Status)
			if err != nil {
				if err == db.ErrOrderAlreadyDelivered || err == db.ErrOrderAlreadyPickedUp || err == db.ErrOrderStatusAlreadySet {
					d.Reject(false)
					return nil
				}
				logging.GetLogger(c).Error(err.Error())
				if errors.Is(err, sql.ErrNoRows) {
					d.Reject(false)
					return nil
				}
				d.Nack(false, true)
				return nil
			}
			d.Ack(false)
		}
		return nil
	})

	ret := &Context{
		config:       c,
		customerRepo: customerRepo,
		proverRepo:   providerRepo,
		ordersRepo:   orderRepo,
		smsMq:        smsMQ,
		statusMQ:     statMQ,
		cacnel:       cancel,
		smsCh:        smsCh,
		smsErrCh:     smsErrCh,
		statusCh:     statCh,
		statusErrCh:  statErrCh,
	}
	return ret, nil
}

func (c *Context) CreateOrder(ctx context.Context, ord *db.Order) (*db.Order, error) {
	order, err := c.ordersRepo.CreateOrder(ctx, ord)
	if err != nil {
		return nil, err
	}
	prov, err := c.proverRepo.GetProviderByID(ctx, order.ProviderID)
	if err != nil {
		return nil, err
	}
	go func() {
		err := providerclient.NotifiyProvider(prov.PickupEP, &providerclient.Order{
			ID:            order.ID,
			Status:        order.Status,
			SenderPhone:   order.SenderPhone,
			ReceiverPhone: order.ReceiverPhone,
			Address:       order.Address,
			CreatedAt:     time.Now(),
		})
		if err != nil {
			logging.GetLogger(c.config).Error(err.Error())
		}
	}()
	return order, nil
}

func (c *Context) PickupOrder(ctx context.Context, id uint64) error {
	res, err := c.ordersRepo.SetOrderPickedUp(ctx, id)
	if err != nil {
		return err
	}
	if !res.PickupSMSSent {
		go func() {
			msg := &mq.SMSMsg{
				OrderID:     res.OrderID,
				PhoneNumber: res.ReceiverPhone,
				CreatedAt:   time.Now(),
			}
			select {
			case <-c.cacnel:
				return
			case c.smsCh <- msg:
				return

			}
		}()
	}

	go func() {
		msg := &mq.StatusMsg{
			OrderID:   res.OrderID,
			Status:    res.Status,
			StatusEP:  res.StatusEP,
			CreatedAt: time.Now(),
		}
		select {
		case <-c.cacnel:
			return
		case c.statusCh <- msg:
			return
		}
	}()
	go func() {
		logging.GetLogger(c.config).Info("notifying provider", zap.Uint64("order_id", res.OrderID), zap.String("status", res.Status))
		err := providerclient.NotifiyProvider(res.NotifyEP, &providerclient.Order{
			ID:     res.OrderID,
			Status: res.Status,
		})
		if err != nil {
			logging.GetLogger(c.config).Error(err.Error())
		}
	}()
	return nil
}

func (c *Context) CreateCustomer(ctx context.Context, customer *db.Customer) (*db.Customer, error) {
	res, err := c.customerRepo.CreateCustomer(ctx, customer)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Context) GetAllCustomers(ctx context.Context) ([]db.Customer, error) {
	res, err := c.customerRepo.GetAllCustomers(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Context) CreateProvider(ctx context.Context, provider *db.Provider) (*db.Provider, error) {
	res, err := c.proverRepo.CreateProvider(ctx, provider)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Context) GetAllProviders(ctx context.Context) ([]db.Provider, error) {
	res, err := c.proverRepo.GetAllProviders(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Context) GetOrderByID(ctx context.Context, id uint64) (*db.Order, error) {
	res, err := c.ordersRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Context) GetAverageDeliveryByProvider(ctx context.Context) ([]db.DeliveryAvg, error) {
	return c.proverRepo.GetAverageDelivery(ctx)
}
