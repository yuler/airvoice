package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseInbound_PreservesNewlines(t *testing.T) {
	raw := []byte(`{"type":"text","id":"abc","content":"你好\n世界","ts":1}`)
	msg, err := ParseInbound(raw)
	if err != nil {
		t.Fatalf("ParseInbound error: %v", err)
	}
	if msg.Type != "text" {
		t.Errorf("Type: got %q, want %q", msg.Type, "text")
	}
	if msg.ID != "abc" {
		t.Errorf("ID: got %q, want %q", msg.ID, "abc")
	}
	if msg.TS != 1 {
		t.Errorf("TS: got %d, want 1", msg.TS)
	}
	// JSON \n decodes to a literal newline character
	if !strings.Contains(msg.Content, "\n") {
		t.Errorf("Content should contain a newline; got %q", msg.Content)
	}
	if msg.Content != "你好\n世界" {
		t.Errorf("Content: got %q, want %q", msg.Content, "你好\n世界")
	}
}

func TestOutbound_RoundTrip(t *testing.T) {
	original := Outbound{
		Type:    "ack",
		Host:    "myhost",
		Version: "1.0.0",
		ID:      "xyz",
		OK:      true,
		Message: "all good",
	}

	data, err := original.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}

	var decoded Outbound
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	if decoded != original {
		t.Errorf("round-trip mismatch:\n  got  %+v\n  want %+v", decoded, original)
	}
}

func TestOutbound_OmitEmpty(t *testing.T) {
	// Only required field set; omitempty fields should be absent from JSON
	o := Outbound{Type: "ping"}
	data, err := o.Bytes()
	if err != nil {
		t.Fatalf("Bytes() error: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if _, ok := m["host"]; ok {
		t.Error("expected 'host' to be omitted when empty")
	}
	if _, ok := m["ok"]; ok {
		t.Error("expected 'ok' to be omitted when false (zero value)")
	}
	if m["type"] != "ping" {
		t.Errorf("type: got %v, want \"ping\"", m["type"])
	}
}

func TestParseInbound_UnknownFields(t *testing.T) {
	// Extra fields in JSON should be silently ignored
	raw := []byte(`{"type":"text","content":"hello","unknown_field":"ignored"}`)
	msg, err := ParseInbound(raw)
	if err != nil {
		t.Fatalf("ParseInbound error on unknown fields: %v", err)
	}
	if msg.Type != "text" || msg.Content != "hello" {
		t.Errorf("unexpected result: %+v", msg)
	}
}
