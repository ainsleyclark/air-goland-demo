package store

import "encoding/json"

// Event is a webhook event sent by a payment provider.
type Event struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
	Provider string          `json:"-"`
}
