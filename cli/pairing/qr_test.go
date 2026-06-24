package pairing

import "testing"

func TestPrintQR(t *testing.T) {
	// Call PrintQR with a simple payload to ensure it doesn't panic
	payload := []byte("test-payload")
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintQR panicked: %v", r)
		}
	}()
	PrintQR(payload)
}
