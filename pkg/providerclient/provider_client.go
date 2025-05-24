package providerclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrProverNotify = errors.New("error notifying provider")

type Order struct {
	ID            uint64    `json:"order_id"`
	Status        string    `json:"status"`
	SenderPhone   string    `json:"sender_phone"`
	ReceiverPhone string    `json:"receiver_phone"`
	Address       string    `json:"address"`
	CreatedAt     time.Time `json:"created_at"`
}

type ProviderResponse struct {
	Message string `json:"message"`
	OrderID uint64 `json:"order_id"`
}

type ReqBoy struct {
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func NotifiyProvider(providerUrl string, ord *Order) error {
	rb := ReqBoy{
		Status: ord.Status,
	}
	providerUrl = fmt.Sprintf("%s/%d", providerUrl, ord.ID)
	js, err := json.Marshal(&rb)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(js)
	req, err := http.NewRequest(http.MethodPost, providerUrl, bodyReader)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrProverNotify
	}

	return nil

}
