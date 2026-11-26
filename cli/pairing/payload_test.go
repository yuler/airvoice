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

func TestPayloadV2MarshalAndURI(t *testing.T) {
	key := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" // 32 zero bytes in raw base64url

	pNoAuth := &PayloadV2{
		Version: VersionRelay,
		Mode:    ModeRelay,
		WS:      "wss://relay.airvoice.yuler.cc/ws",
		Key:     key,
	}
	raw, err := pNoAuth.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	wantJSON := `{"v":2,"mode":"relay","ws":"wss://relay.airvoice.yuler.cc/ws","key":"` + key + `"}`
	if string(raw) != wantJSON {
		t.Fatalf("JSON = %s, want %s", raw, wantJSON)
	}

	uri, err := pNoAuth.PairingURI()
	if err != nil {
		t.Fatalf("PairingURI: %v", err)
	}
	wantURI := "airvoice://pair?v=2&mode=relay&ws=wss%3A%2F%2Frelay.airvoice.yuler.cc%2Fws&key=" + key
	if uri != wantURI {
		t.Fatalf("URI = %q, want %q", uri, wantURI)
	}

	pWithAuth := &PayloadV2{
		Version: VersionRelay,
		Mode:    ModeRelay,
		WS:      "wss://custom.example.com/ws",
		Key:     key,
		Auth:    "my secret+key",
	}
	rawAuth, err := pWithAuth.Marshal()
	if err != nil {
		t.Fatalf("Marshal with auth: %v", err)
	}
	if !bytes.Contains(rawAuth, []byte(`"auth":"my secret+key"`)) {
		t.Fatalf("expected auth field in JSON: %s", rawAuth)
	}

	uriAuth, err := pWithAuth.PairingURI()
	if err != nil {
		t.Fatalf("PairingURI with auth: %v", err)
	}
	if !bytes.Contains([]byte(uriAuth), []byte("&auth=my%20secret%2Bkey")) {
		t.Fatalf("unexpected encoded auth in URI: %q", uriAuth)
	}

	if _, err := BuildPairingURI("wss://relay.airvoice.yuler.cc/ws", "bad-key", ""); err == nil {
		t.Fatal("expected error for invalid relay key")
	}
}

func TestFprintRelayPairing(t *testing.T) {
	key := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	var buf bytes.Buffer
	uri, err := fprintRelayPairing(&buf, "wss://relay.airvoice.yuler.cc/ws", key, "", "", "0.5.1")
	if err != nil {
		t.Fatalf("fprintRelayPairing: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"Airvoice 0.5.1",
		"Relay: wss://relay.airvoice.yuler.cc/ws (E2E)",
		"Link: " + uri,
	} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}
