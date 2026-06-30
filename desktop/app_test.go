package main

import (
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
