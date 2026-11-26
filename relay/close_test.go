package relay

import (
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWriteLoopAlwaysSendsExplicitCloseCode(t *testing.T) {
	cfg := DefaultConfig()
	_, baseURL := newTestServer(t, cfg)
	keys := newDerivedPair(t)

	prev := mustDialPeer(t, baseURL, keys.RoomID, RoleHost, keys.HostAuth, "", nil)
	_ = readJSON(t, prev) // initial peer_status

	// Repeat replacement enough times to catch any select race between closeCh and done.
	for i := range 30 {
		next := mustDialPeer(t, baseURL, keys.RoomID, RoleHost, keys.HostAuth, "", nil)
		_ = readJSON(t, next) // initial peer_status

		_ = prev.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, _, err := prev.ReadMessage()
		if err == nil {
			t.Fatalf("iteration %d: expected close error on replaced peer, got nil", i)
		}
		var closeErr *websocket.CloseError
		if !errors.As(err, &closeErr) || closeErr.Code != CloseCodeReplaced {
			t.Fatalf("iteration %d: expected close code %d, got %v", i, CloseCodeReplaced, err)
		}
		_ = prev.Close()
		prev = next
	}
}
