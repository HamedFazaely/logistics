package db

import "errors"

var (
	ErrOrderAlreadyPickedUp  = errors.New("order already picked up")
	ErrOrderAlreadyDelivered = errors.New("order already delivered")
	ErrOrderExpired          = errors.New("order pick up time is due")
	ErrOrderStatusAlreadySet = errors.New("order status already set")
	ErrPickupSMSAlreadySent = errors.New("pickup sms already sent")
)
