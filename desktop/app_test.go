package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
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
	t.Logf("Generated QR Code PNG bounds: %v (Width: %d, Height: %d)", bounds, bounds.Dx(), bounds.Dy())

	// Analyze unique colors
	colorCount := make(map[string]int)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			colorCount[fmt.Sprintf("%v", img.At(x, y))] = colorCount[fmt.Sprintf("%v", img.At(x, y))] + 1
		}
	}
	t.Logf("Unique colors count: %d", len(colorCount))
	for col, count := range colorCount {
		t.Logf("Color: %s, Count: %d", col, count)
	}
}


