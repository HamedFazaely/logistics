package db

import "time"

const (
	Pending      = "0"
	ProviderSeen = "1"
	PickedUp     = "2"
	InProgress   = "3"
	Delivered    = "4"
)

type Customer struct {
	ID        uint64    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Provider struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	StatusEP  string    `json:"status_ep"`
	PickupEP  string    `json:"pickup_ep"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Order struct {
	ID            uint64     `json:"id"`
	CutomerID     uint16     `json:"customer_id"`
	ProviderID    uint64     `json:"provider_id"`
	Status        string     `json:"status"`
	SenderPhone   string     `json:"sender_phone"`
	ReceiverPhone string     `json:"receiver_phone"`
	Address       string     `json:"address"`
	PickupSMSSent bool       `json:"-"`
	PickupTime    *time.Time `json:"-"`
	DeliveryTime  *time.Time `json:"-"`
	CreatedAt     time.Time  `json:"-"`
	UpdatedAt     time.Time  `json:"-"`
}

type OrderWithProvider struct {
	OrderID       uint64
	PickupSMSSent bool
	Status        string
	StatusEP      string
	ReceiverPhone string
	NotifyEP      string
}

type DeliveryAvg struct {
	Avg          float64 `json:"average_delivery"`
	ProviderName string  `json:"provider_name"`
}
