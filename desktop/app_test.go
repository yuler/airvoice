package main

import (
	"bytes"
	"encoding/base64"
	"image/png"
	"os"
	"testing"
)

func TestNewApp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "airvoice-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	app, err := NewApp()
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	if app.token == "" {
		t.Error("token should not be empty")
	}
	if app.port != 7383 {
		t.Errorf("port = %d, want 7383", app.port)
	}
}

func TestGetSettings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "airvoice-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	app, _ := NewApp()
	settings := app.GetSettings()

	if settings.Port != 7383 {
		t.Errorf("settings.Port = %d, want 7383", settings.Port)
	}
	if settings.Language != "zh-CN" {
		t.Errorf("settings.Language = %s, want zh-CN", settings.Language)
	}
}

func TestGetQRCode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "airvoice-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", origUserProfile)

	app, _ := NewApp()
	qrCodeStr, err := app.GetQRCode()
	if err != nil {
		t.Fatalf("GetQRCode() error = %v", err)
	}

	if qrCodeStr == "" {
		t.Fatal("GetQRCode() returned empty string")
	}

	// Verify it's a data URL
	prefix := "data:image/png;base64,"
	if len(qrCodeStr) <= len(prefix) || qrCodeStr[:len(prefix)] != prefix {
		t.Fatalf("invalid QR code data URL prefix: %s", qrCodeStr)
	}

	rawB64 := qrCodeStr[len(prefix):]
	data, err := base64.StdEncoding.DecodeString(rawB64)
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to decode PNG image: %v", err)
	}

	bounds := img.Bounds()

	// Analyze unique colors and assert correct structure.
	var blackCount, whiteCount int
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0 && g == 0 && b == 0 {
				blackCount++
			} else if r == 0xFFFF && g == 0xFFFF && b == 0xFFFF {
				whiteCount++
			}
		}
	}

	if bounds.Dx() < 100 || bounds.Dy() < 100 {
		t.Errorf("QR code dimensions are too small: %dx%d", bounds.Dx(), bounds.Dy())
	}
	if blackCount == 0 || whiteCount == 0 {
		t.Error("QR code should contain both black and white pixels")
	}
	// A correct QR code of scale=8 will have a substantial amount of both black and white pixels.
	// If it was bugged (1-pixel scale), blackCount would be extremely small (e.g. < 1000).
	if blackCount < 5000 {
		t.Errorf("QR code black pixels count is too low (%d), indicating scaling issue", blackCount)
	}
}


