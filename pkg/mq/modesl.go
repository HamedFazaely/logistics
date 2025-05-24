package mq

import "time"

type SMSMsg struct {
	OrderID     uint64    `json:"order_id"`
	PhoneNumber string    `json:"phone_number"`
	CreatedAt   time.Time `json:"created_at"`
}

type SMSPubError struct {
	Err error
	Msg *SMSMsg
}

type StatusMsg struct {
	OrderID   uint64    `json:"order_id"`
	Status    string    `json:"status"`
	StatusEP  string    `json:"status_ep"`
	CreatedAt time.Time `json:"created_at"`
}

type StatusPubError struct {
	Err error
	Msg *StatusMsg
}
