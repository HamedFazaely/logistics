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

type SMSPublisherConsumer interface {
	AMQPConsumer
	StartPublisher(msg <-chan *SMSMsg, errPub chan<- *SMSPubError)
}

type SMS struct {
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

func NewSMS(conf *conf.Configuration, cancel <-chan struct{}) (*SMS, error) {
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

	ret := &SMS{
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

func (s *SMS) StartPublisher(msg <-chan *SMSMsg, errPub chan<- *SMSPubError) {

	go func() {
		q, err := s.pubCh.QueueDeclare(
			s.config.SMSQueueName, // name
			false,                 // durable
			false,                 // delete when unused
			false,                  // exclusive
			false,                 // no-wait
			nil,                   // arguments
		)
		if err != nil {
			logging.GetLogger(s.config).Error(err.Error())
			return
		}
	loop:
		for {
			select {
			case <-s.cancel:
				logging.GetLogger(s.config).Info("closing sms mq connection due to cancel")
				s.pubCh.Close()
				s.pubConn.Close()
				break loop
			case m := <-msg:
				body, err := json.Marshal(m)
				if err != nil {
					logging.GetLogger(s.config).Error(err.Error())
				}

				logging.GetLogger(s.config).Info("sending message to sms q", zap.Any("msg", m))
				ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)

				err = s.pubCh.PublishWithContext(ctx,
					"",     // exchange
					q.Name, // routing key
					false,  // mandatory
					false,  // immediate
					amqp.Publishing{
						ContentType: "application/json",
						Body:        body,
					})
				if err != nil {
					logging.GetLogger(s.config).Error(err.Error())
					errPub <- &SMSPubError{err, m}
				}
				cancelFunc()
			case e := <-s.pubCloseNotify:
				if e != nil {
					logging.GetLogger(s.config).Info("trying to reconnect to rabbitMQ")
					ch := make(chan *amqp.Error)
					conn, channel, err := createAMQPConn(s.config, ch)
					if err != nil {
						logging.GetLogger(s.config).Error(err.Error())
						return //Todo: retry reconnect
					}
					s.mu.Lock()
					s.pubCloseNotify = ch
					s.pubConn = conn
					s.pubCh = channel
					s.mu.Unlock()

				}
			}
		}
	}()

}

func (s *SMS) StartConsume(onReceive func(amqp.Delivery, <-chan struct{}) error) {
	go func() {
		q, err := s.consumeCh.QueueDeclare(
			s.config.SMSReplyQueueName, // name
			false,                      // durable
			false,                      // delete when unused
			false,                       // exclusive
			false,                      // no-wait
			nil,                        // arguments
		)
		if err != nil {
			logging.GetLogger(s.config).Error(err.Error())
			return
		}
		msgs, err := s.consumeCh.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			logging.GetLogger(s.config).Error(err.Error())
			return
		}
	loop:
		for {
			select {
			case <-s.cancel:
				logging.GetLogger(s.config).Info("closing sms consumer connections due to cancel")
				s.consumeCh.Close()
				s.consumeConn.Close()
				break loop
			case d, ok := <-msgs:
				if ok {
					logging.GetLogger(s.config).Info("reply sms message arrived")
					go onReceive(d, s.cancel)
				}

			case conErr := <-s.consumeCloseNotify:
				if conErr != nil {
					logging.GetLogger(s.config).Info("trying to reconnect to rabbitmq")
					ch := make(chan *amqp.Error)
					conn, channel, err := createAMQPConn(s.config, ch)
					if err != nil {
						logging.GetLogger(s.config).Error(err.Error())
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
						logging.GetLogger(s.config).Error(err.Error())
						return
					}
					s.mu.Lock()
					s.consumeCloseNotify = ch
					s.consumeConn = conn
					s.consumeCh = channel
					s.mu.Unlock()
				}

			}
		}
	}()
}
