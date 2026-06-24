package pairing

import "encoding/json"

// Payload represents the pairing information sent to the peer.
type Payload struct {
	Version int    `json:"v"`
	WS      string `json:"ws"`
	Token   string `json:"token"`
}

// Marshal serializes the Payload to a JSON byte slice.
func (p *Payload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}
