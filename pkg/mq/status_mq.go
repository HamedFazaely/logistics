package mq

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.com/Hamed1984/logistics/pkg/conf"
	"gitlab.com/Hamed1984/logistics/pkg/logging"
	"go.uber.org/zap"
)

type StatusPublisherReceiver interface {
	AMQPConsumer
	StartPublisher(msg <-chan *StatusMsg, errPub chan<- *StatusPubError)
}

type StatusMQ struct {
	pubConn            *amqp.Connection
	pubCh              *amqp.Channel
	pubCloseNotify     chan *amqp.Error
	consumeConn        *amqp.Connection
	consumeCh          *amqp.Channel
	consumeCloseNotify chan *amqp.Error
	config             *conf.Configuration
	cancel             <-chan struct{}
	mu                 sync.Mutex
}

func NewStatusMQ(conf *conf.Configuration, cancel <-chan struct{}) (*StatusMQ, error) {
	pcn := make(chan *amqp.Error)
	pConn, pCh, err := createAMQPConn(conf, pcn)
	if err != nil {
		return nil, err
	}

	consumeNotifyClose := make(chan *amqp.Error)
	cConn, cCh, err := createAMQPConn(conf, consumeNotifyClose)
	if err != nil {
		return nil, err
	}

	ret := &StatusMQ{
		pubConn:            pConn,
		pubCh:              pCh,
		config:             conf,
		cancel:             cancel,
		pubCloseNotify:     pcn,
		consumeConn:        cConn,
		consumeCh:          cCh,
		consumeCloseNotify: consumeNotifyClose,
	}
	return ret, nil

}

func (o *StatusMQ) StartPublisher(msg <-chan *StatusMsg, errPub chan<- *StatusPubError) {
	go func() {
		q, err := o.pubCh.QueueDeclare(
			o.config.StatusQueueName, // name
			false,                    // durable
			false,                    // delete when unused
			false,                     // exclusive
			false,                    // no-wait
			nil,                      // arguments
		)
		if err != nil {
			logging.GetLogger(o.config).Error(err.Error())
			return
		}
	loop:
		for {
			select {
			case <-o.cancel:
				logging.GetLogger(o.config).Info("closing status mq connections due to canel")
				o.pubCh.Close()
				o.pubConn.Close()
				break loop
			case m := <-msg:
				body, err := json.Marshal(m)
				if err != nil {
					logging.GetLogger(o.config).Error(err.Error())
				}
				logging.GetLogger(o.config).Info("publishing status message", zap.Any("msg", m))
				ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)

				err = o.pubCh.PublishWithContext(ctx,
					"",     // exchange
					q.Name, // routing key
					false,  // mandatory
					false,  // immediate
					amqp.Publishing{
						ContentType: "application/json",
						Body:        body,
					})
				if err != nil {
					logging.GetLogger(o.config).Error(err.Error())
					errPub <- &StatusPubError{err, m}
				}
				cancelFunc()
			case e := <-o.pubCloseNotify:
				if e != nil {
					logging.GetLogger(o.config).Info("trying to reconnect to tabbitmq")
					ch := make(chan *amqp.Error)
					conn, channel, err := createAMQPConn(o.config, ch)
					if err != nil {
						logging.GetLogger(o.config).Error(err.Error())
						return //Todo: retry reconnect
					}
					o.mu.Lock()
					o.pubCloseNotify = ch
					o.pubConn = conn
					o.pubCh = channel
					o.mu.Unlock()

				}
			}
		}
	}()
}

func (o *StatusMQ) StartConsume(onReceive func(amqp.Delivery, <-chan struct{}) error) {
	go func() {
		q, err := o.consumeCh.QueueDeclare(
			o.config.StatusReplyQueueName, // name
			false,                         // durable
			false,                         // delete when unused
			false,                          // exclusive
			false,                         // no-wait
			nil,                           // arguments
		)
		if err != nil {
			logging.GetLogger(o.config).Error(err.Error())
			return
		}
		msgs, err := o.consumeCh.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			logging.GetLogger(o.config).Error(err.Error())
			return
		}
	loop:
		for {
			select {
			case <-o.cancel:
				logging.GetLogger(o.config).Info("closing status consumer connections due to cancel")
				o.consumeCh.Close()
				o.consumeConn.Close()
				break loop
			case d, ok := <-msgs:
				if ok {
					logging.GetLogger(o.config).Info("status reply message arrived")
					go onReceive(d, o.cancel)
				}

			case conErr := <-o.consumeCloseNotify:
				if conErr != nil {
					logging.GetLogger(o.config).Info("trying to reconnect to rabbitmq")
					ch := make(chan *amqp.Error)
					conn, channel, err := createAMQPConn(o.config, ch)
					if err != nil {
						logging.GetLogger(o.config).Error(err.Error())
						return //Todo: retry reconnect
					}
					msgs, err = channel.Consume(
						q.Name, // queue
						"",     // consumer
						false,  // auto-ack
						false,  // exclusive
						false,  // no-local
						false,  // no-wait
						nil,    // args
					)
					if err != nil {
						logging.GetLogger(o.config).Error(err.Error())
						return
					}
					o.mu.Lock()
					o.consumeCloseNotify = ch
					o.consumeConn = conn
					o.consumeCh = channel
					o.mu.Unlock()
				}

			}
		}
	}()
}
