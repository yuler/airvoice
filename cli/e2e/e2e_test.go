package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type testVectorFile struct {
	Description string            `json:"description"`
	RelayKey    string            `json:"relayKey"`
	Derived     derivedVector     `json:"derived"`
	Frames      []frameTestVector `json:"frames"`
}

type derivedVector struct {
	RoomID     string `json:"roomId"`
	HostAuth   string `json:"hostAuth"`
	ClientAuth string `json:"clientAuth"`
	KeyC2H     string `json:"keyC2H"`
	KeyH2C     string `json:"keyH2C"`
}

type frameTestVector struct {
	Name      string    `json:"name"`
	Direction Direction `json:"direction"`
	SID       string    `json:"sid"`
	Seq       uint64    `json:"seq"`
	IV        string    `json:"iv"`
	AAD       string    `json:"aad"`
	Plaintext string    `json:"plaintext"`
	Envelope  Envelope  `json:"envelope"`
}

func TestGenerateAndParseKey(t *testing.T) {
	k1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	k2, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if len(k1) != 43 {
		t.Fatalf("expected 43-char key, got %d (%q)", len(k1), k1)
	}
	if k1 == k2 {
		t.Fatal("expected distinct random keys")
	}

	raw, err := ParseKey(k1)
	if err != nil {
		t.Fatalf("ParseKey failed: %v", err)
	}
	if len(raw) != KeySize {
		t.Fatalf("expected %d bytes, got %d", KeySize, len(raw))
	}

	for _, bad := range []string{"", "short", k1 + "=", "invalid!@#"} {
		if _, err := ParseKey(bad); !errors.Is(err, ErrInvalidKey) {
			t.Errorf("ParseKey(%q): expected ErrInvalidKey, got %v", bad, err)
		}
	}
}

func TestDeriveKeysDeterministic(t *testing.T) {
	raw := make([]byte, KeySize)
	for i := range raw {
		raw[i] = byte(i + 1)
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)

	d1, err := DeriveKeysFromBase64(encoded)
	if err != nil {
		t.Fatalf("DeriveKeysFromBase64 failed: %v", err)
	}
	d2, err := DeriveKeys(raw)
	if err != nil {
		t.Fatalf("DeriveKeys failed: %v", err)
	}

	if d1.RoomID != d2.RoomID || len(d1.RoomID) != 22 {
		t.Fatalf("unexpected roomId: %q vs %q", d1.RoomID, d2.RoomID)
	}
	if d1.HostAuth != d2.HostAuth || len(d1.HostAuth) != 43 {
		t.Fatalf("unexpected hostAuth: %q", d1.HostAuth)
	}
	if d1.ClientAuth != d2.ClientAuth || len(d1.ClientAuth) != 43 {
		t.Fatalf("unexpected clientAuth: %q", d1.ClientAuth)
	}
	if d1.HostAuth == d1.ClientAuth {
		t.Fatal("hostAuth and clientAuth must differ")
	}
	if !bytes.Equal(d1.KeyC2H, d2.KeyC2H) || len(d1.KeyC2H) != KeySize {
		t.Fatal("keyC2H mismatch")
	}
	if !bytes.Equal(d1.KeyH2C, d2.KeyH2C) || len(d1.KeyH2C) != KeySize {
		t.Fatal("keyH2C mismatch")
	}
	if bytes.Equal(d1.KeyC2H, d1.KeyH2C) {
		t.Fatal("keyC2H and keyH2C must differ")
	}
}

func TestSealOpenAndTampering(t *testing.T) {
	keyStr, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	keys, err := DeriveKeysFromBase64(keyStr)
	if err != nil {
		t.Fatalf("DeriveKeysFromBase64: %v", err)
	}
	sid, err := NewSessionID()
	if err != nil {
		t.Fatalf("NewSessionID: %v", err)
	}
	plaintext := []byte(`{"type":"text","id":"1","content":"hello 世界 🎙️"}`)

	env, err := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid, 1, plaintext)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	got, err := Open(keys.KeyC2H, DirC2H, keys.RoomID, env)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("plaintext mismatch: got %q want %q", got, plaintext)
	}

	// Wrong direction key
	if _, err := Open(keys.KeyH2C, DirC2H, keys.RoomID, env); !errors.Is(err, ErrAuthFailed) {
		t.Errorf("wrong key: got %v, want ErrAuthFailed", err)
	}
	// Wrong direction AAD
	if _, err := Open(keys.KeyC2H, DirH2C, keys.RoomID, env); !errors.Is(err, ErrAuthFailed) {
		t.Errorf("wrong dir: got %v, want ErrAuthFailed", err)
	}
	// Wrong roomId AAD
	if _, err := Open(keys.KeyC2H, DirC2H, "other-room-id-12345678", env); !errors.Is(err, ErrAuthFailed) {
		t.Errorf("wrong roomId: got %v, want ErrAuthFailed", err)
	}
	// Tampered seq
	badSeq := env
	badSeq.Seq = 2
	if _, err := Open(keys.KeyC2H, DirC2H, keys.RoomID, badSeq); !errors.Is(err, ErrAuthFailed) {
		t.Errorf("tampered seq: got %v, want ErrAuthFailed", err)
	}
	// Tampered ciphertext
	ctRaw, _ := base64.RawURLEncoding.DecodeString(env.CT)
	ctRaw[0] ^= 0xff
	badCT := env
	badCT.CT = base64.RawURLEncoding.EncodeToString(ctRaw)
	if _, err := Open(keys.KeyC2H, DirC2H, keys.RoomID, badCT); !errors.Is(err, ErrAuthFailed) {
		t.Errorf("tampered ct: got %v, want ErrAuthFailed", err)
	}
}

func TestSessionGuardHelloGatingAndReplayProtection(t *testing.T) {
	keyStr, _ := GenerateKey()
	keys, _ := DeriveKeysFromBase64(keyStr)

	now := time.Unix(1700000000, 0)
	guard, err := NewSessionGuard(keys.KeyC2H, DirC2H, keys.RoomID, WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatalf("NewSessionGuard: %v", err)
	}

	sid1, _ := NewSessionID()

	// 1. Unauthenticated SID sending text at seq 1 must be rejected.
	textAt1, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 1, []byte(`{"type":"text","id":"1","content":"sneaky"}`))
	if _, err := guard.Open(textAt1); !errors.Is(err, ErrHelloRequired) {
		t.Fatalf("expected ErrHelloRequired for text at seq 1, got %v", err)
	}

	// 2. Unauthenticated SID sending text at seq 2 must be rejected.
	textAt2, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 2, []byte(`{"type":"text","id":"1","content":"sneaky"}`))
	if _, err := guard.Open(textAt2); !errors.Is(err, ErrHelloRequired) {
		t.Fatalf("expected ErrHelloRequired for text at seq 2, got %v", err)
	}

	// 3. Stale hello.ts (> 5 min old) must be rejected.
	staleHello, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 1, []byte(fmt.Sprintf(`{"type":"hello","ts":%d}`, now.Add(-6*time.Minute).Unix())))
	if _, err := guard.Open(staleHello); !errors.Is(err, ErrStaleHello) {
		t.Fatalf("expected ErrStaleHello for old ts, got %v", err)
	}

	// 4. Future hello.ts (> 5 min ahead) must be rejected.
	futureHello, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 1, []byte(fmt.Sprintf(`{"type":"hello","ts":%d}`, now.Add(6*time.Minute).Unix())))
	if _, err := guard.Open(futureHello); !errors.Is(err, ErrStaleHello) {
		t.Fatalf("expected ErrStaleHello for future ts, got %v", err)
	}

	// 5. Missing hello.ts must be rejected.
	noTSHello, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 1, []byte(`{"type":"hello"}`))
	if _, err := guard.Open(noTSHello); !errors.Is(err, ErrStaleHello) {
		t.Fatalf("expected ErrStaleHello for missing ts, got %v", err)
	}

	// 6. Valid hello at seq 1 opens session sid1.
	validHello, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 1, []byte(fmt.Sprintf(`{"type":"hello","device":"iPhone","ts":%d}`, now.Unix())))
	if _, err := guard.Open(validHello); err != nil {
		t.Fatalf("valid hello failed: %v", err)
	}

	// 7. Replaying seq 1 must fail.
	if _, err := guard.Open(validHello); !errors.Is(err, ErrReplayedSequence) {
		t.Fatalf("expected ErrReplayedSequence, got %v", err)
	}

	// 8. Text at seq 2 succeeds; seq 2 replay fails; seq 4 succeeds; out-of-order seq 3 fails.
	if _, err := guard.Open(textAt2); err != nil {
		t.Fatalf("text at seq 2 failed: %v", err)
	}
	if _, err := guard.Open(textAt2); !errors.Is(err, ErrReplayedSequence) {
		t.Fatalf("expected ErrReplayedSequence on seq 2 replay, got %v", err)
	}
	textAt4, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 4, []byte(`{"type":"text","id":"4","content":"ok"}`))
	if _, err := guard.Open(textAt4); err != nil {
		t.Fatalf("text at seq 4 failed: %v", err)
	}
	textAt3, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 3, []byte(`{"type":"text","id":"3","content":"late"}`))
	if _, err := guard.Open(textAt3); !errors.Is(err, ErrReplayedSequence) {
		t.Fatalf("expected ErrReplayedSequence on seq 3 < 4, got %v", err)
	}

	// 9. Bad ciphertext on a new sid must NOT poison active sid1.
	sidBad, _ := NewSessionID()
	badEnv, _ := Seal(keys.KeyH2C, DirC2H, keys.RoomID, sidBad, 1, []byte(fmt.Sprintf(`{"type":"hello","ts":%d}`, now.Unix())))
	if _, err := guard.Open(badEnv); !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("expected ErrAuthFailed, got %v", err)
	}
	textAt5, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 5, []byte(`{"type":"text","id":"5","content":"still alive"}`))
	if _, err := guard.Open(textAt5); err != nil {
		t.Fatalf("sid1 should remain active after failed foreign frame: %v", err)
	}

	// 10. Valid new sid2 retires sid1 permanently.
	sid2, _ := NewSessionID()
	hello2, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid2, 1, []byte(fmt.Sprintf(`{"type":"hello","ts":%d}`, now.Unix())))
	if _, err := guard.Open(hello2); err != nil {
		t.Fatalf("valid hello on sid2 failed: %v", err)
	}
	textOnRetired, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sid1, 6, []byte(`{"type":"text","id":"6","content":"old"}`))
	if _, err := guard.Open(textOnRetired); !errors.Is(err, ErrReplayedSession) {
		t.Fatalf("expected ErrReplayedSession for retired sid1, got %v", err)
	}

	// 11. RetireCurrent retires sid2 so even a fresh hello on sid2 is rejected.
	guard.RetireCurrent()
	if _, err := guard.Open(hello2); !errors.Is(err, ErrReplayedSession) {
		t.Fatalf("expected ErrReplayedSession after RetireCurrent, got %v", err)
	}
}

func TestSessionGuardRetiredRingEviction(t *testing.T) {
	keyStr, _ := GenerateKey()
	keys, _ := DeriveKeysFromBase64(keyStr)
	now := time.Unix(1700000000, 0)

	guard, err := NewSessionGuard(keys.KeyC2H, DirC2H, keys.RoomID, WithClock(func() time.Time { return now }), WithRetiredLimit(3))
	if err != nil {
		t.Fatalf("NewSessionGuard: %v", err)
	}

	sids := make([]string, 5)
	for i := range sids {
		sids[i], _ = NewSessionID()
		env, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sids[i], 1, []byte(fmt.Sprintf(`{"type":"hello","ts":%d}`, now.Unix())))
		if _, err := guard.Open(env); err != nil {
			t.Fatalf("open sid[%d]: %v", i, err)
		}
	}
	// Active is sids[4]; retired ring of size 3 holds sids[1], sids[2], sids[3]; sids[0] was evicted.
	for _, idx := range []int{1, 2, 3} {
		env, _ := Seal(keys.KeyC2H, DirC2H, keys.RoomID, sids[idx], 1, []byte(fmt.Sprintf(`{"type":"hello","ts":%d}`, now.Unix())))
		if _, err := guard.Open(env); !errors.Is(err, ErrReplayedSession) {
			t.Fatalf("sid[%d] should still be retired, got %v", idx, err)
		}
	}
	if len(guard.retired) != 3 {
		t.Fatalf("expected retired ring len 3, got %d", len(guard.retired))
	}
}

func TestSenderRoundTrip(t *testing.T) {
	keyStr, _ := GenerateKey()
	keys, _ := DeriveKeysFromBase64(keyStr)
	now := time.Unix(1700000000, 0)

	sender, err := NewSender(keys.KeyC2H, DirC2H, keys.RoomID)
	if err != nil {
		t.Fatalf("NewSender: %v", err)
	}
	guard, err := NewSessionGuard(keys.KeyC2H, DirC2H, keys.RoomID, WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatalf("NewSessionGuard: %v", err)
	}

	e1, err := sender.Seal([]byte(fmt.Sprintf(`{"type":"hello","ts":%d}`, now.Unix())))
	if err != nil {
		t.Fatalf("sender.Seal 1: %v", err)
	}
	if e1.Seq != 1 {
		t.Fatalf("expected seq 1, got %d", e1.Seq)
	}
	if _, err := guard.Open(e1); err != nil {
		t.Fatalf("guard.Open 1: %v", err)
	}

	e2, err := sender.Seal([]byte(`{"type":"text","id":"1","content":"hi"}`))
	if err != nil {
		t.Fatalf("sender.Seal 2: %v", err)
	}
	if e2.Seq != 2 || e2.SID != e1.SID {
		t.Fatalf("unexpected seq/sid: %+v", e2)
	}
	if _, err := guard.Open(e2); err != nil {
		t.Fatalf("guard.Open 2: %v", err)
	}

	if err := sender.Reset(); err != nil {
		t.Fatalf("sender.Reset: %v", err)
	}
	e3, err := sender.Seal([]byte(fmt.Sprintf(`{"type":"hello","ts":%d}`, now.Unix())))
	if err != nil {
		t.Fatalf("sender.Seal 3: %v", err)
	}
	if e3.Seq != 1 || e3.SID == e1.SID {
		t.Fatalf("expected new sid and seq 1 after reset: %+v", e3)
	}
	if _, err := guard.Open(e3); err != nil {
		t.Fatalf("guard.Open 3: %v", err)
	}
}

func TestCrossPlatformVectors(t *testing.T) {
	path := filepath.Join("testdata", "vectors.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read vectors.json: %v", err)
	}

	var vf testVectorFile
	if err := json.Unmarshal(data, &vf); err != nil {
		t.Fatalf("parse vectors.json: %v", err)
	}

	keys, err := DeriveKeysFromBase64(vf.RelayKey)
	if err != nil {
		t.Fatalf("DeriveKeysFromBase64: %v", err)
	}
	if keys.RoomID != vf.Derived.RoomID {
		t.Errorf("roomId mismatch: got %q want %q", keys.RoomID, vf.Derived.RoomID)
	}
	if keys.HostAuth != vf.Derived.HostAuth {
		t.Errorf("hostAuth mismatch: got %q want %q", keys.HostAuth, vf.Derived.HostAuth)
	}
	if keys.ClientAuth != vf.Derived.ClientAuth {
		t.Errorf("clientAuth mismatch: got %q want %q", keys.ClientAuth, vf.Derived.ClientAuth)
	}
	if got := base64.RawURLEncoding.EncodeToString(keys.KeyC2H); got != vf.Derived.KeyC2H {
		t.Errorf("keyC2H mismatch: got %q want %q", got, vf.Derived.KeyC2H)
	}
	if got := base64.RawURLEncoding.EncodeToString(keys.KeyH2C); got != vf.Derived.KeyH2C {
		t.Errorf("keyH2C mismatch: got %q want %q", got, vf.Derived.KeyH2C)
	}

	for _, f := range vf.Frames {
		t.Run(f.Name, func(t *testing.T) {
			key := keys.KeyC2H
			if f.Direction == DirH2C {
				key = keys.KeyH2C
			}
			gotAAD := string(AAD(f.Direction, keys.RoomID, f.SID, f.Seq))
			if gotAAD != f.AAD {
				t.Fatalf("AAD mismatch: got %q want %q", gotAAD, f.AAD)
			}
			nonce, err := base64.RawURLEncoding.DecodeString(f.IV)
			if err != nil {
				t.Fatalf("decode iv: %v", err)
			}
			env, err := SealWithNonce(key, f.Direction, keys.RoomID, f.SID, f.Seq, nonce, []byte(f.Plaintext))
			if err != nil {
				t.Fatalf("SealWithNonce: %v", err)
			}
			if env != f.Envelope {
				t.Fatalf("envelope mismatch:\n got: %+v\nwant: %+v", env, f.Envelope)
			}
			pt, err := Open(key, f.Direction, keys.RoomID, f.Envelope)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			if string(pt) != f.Plaintext {
				t.Fatalf("plaintext mismatch: got %q want %q", pt, f.Plaintext)
			}
		})
	}
}
