package ui

type MessageResp struct {
	Msg  string      `json:"message"`
	Data interface{} `json:"data,omitempty"`
}
