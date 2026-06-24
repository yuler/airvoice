package protocol

import "encoding/json"

type Inbound struct {
	Type    string `json:"type"`
	Device  string `json:"device,omitempty"`
	App     string `json:"app,omitempty"`
	ID      string `json:"id,omitempty"`
	Content string `json:"content,omitempty"`
	TS      int64  `json:"ts,omitempty"`
}

type Outbound struct {
	Type    string `json:"type"`
	Host    string `json:"host,omitempty"`
	Version string `json:"version,omitempty"`
	ID      string `json:"id,omitempty"`
	OK      bool   `json:"ok,omitempty"`
	Message string `json:"message,omitempty"`
}

func ParseInbound(data []byte) (Inbound, error) {
	var m Inbound
	return m, json.Unmarshal(data, &m)
}

func (o Outbound) Bytes() ([]byte, error) { return json.Marshal(o) }
