package pairing

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestPayloadMarshal(t *testing.T) {
	p := &Payload{
		Version: 1,
		WS:      "ws://192.168.1.100:8080/ws",
		Token:   "test-token-123",
	}

	data, err := p.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	// Verify it can be unmarshaled back to the same values
	var decoded Payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if decoded.Version != p.Version {
		t.Errorf("expected Version %d, got %d", p.Version, decoded.Version)
	}
	if decoded.WS != p.WS {
		t.Errorf("expected WS %q, got %q", p.WS, decoded.WS)
	}
	if decoded.Token != p.Token {
		t.Errorf("expected Token %q, got %q", p.Token, decoded.Token)
	}

	// Double check the exact JSON output to verify tag naming
	expectedJSON := `{"v":1,"ws":"ws://192.168.1.100:8080/ws","token":"test-token-123"}`
	if !bytes.Equal(data, []byte(expectedJSON)) {
		t.Errorf("expected JSON %s, got %s", expectedJSON, string(data))
	}
}
