//go:build windows

package paste

import (
	"encoding/base64"
	"errors"
	"reflect"
	"testing"
)

type commandCall struct {
	name  string
	stdin string
	args  []string
}

func TestWindowsPaster(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	var calls []commandCall
	runCommand = func(name string, stdin string, args ...string) error {
		calls = append(calls, commandCall{name: name, stdin: stdin, args: args})
		return nil
	}

	t.Run("windowsPaster success", func(t *testing.T) {
		calls = nil
		p := &windowsPaster{}
		if p.Name() != "windows" {
			t.Errorf("expected windows, got %s", p.Name())
		}
		text := "hello world"
		err := p.Paste(text)
		if err != nil {
			t.Fatalf("Paste failed: %v", err)
		}

		encoded := base64.StdEncoding.EncodeToString([]byte(text))
		psCmd := `$ErrorActionPreference = 'Stop'; $b = [Console]::In.ReadToEnd().Trim(); if ($b) { Set-Clipboard -Value ([System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($b))); Start-Sleep -Milliseconds 80; (New-Object -ComObject WScript.Shell).SendKeys('^v') }`

		expected := []commandCall{
			{name: "powershell", stdin: encoded, args: []string{"-NoProfile", "-Command", psCmd}},
		}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("got calls %+v, expected %+v", calls, expected)
		}
	})

	t.Run("windowsPaster empty text", func(t *testing.T) {
		calls = nil
		p := &windowsPaster{}
		err := p.Paste("")
		if err != nil {
			t.Fatalf("Paste failed: %v", err)
		}
		if len(calls) != 0 {
			t.Errorf("expected no commands to be run for empty text, got %d", len(calls))
		}
	})

	t.Run("windowsPaster failure", func(t *testing.T) {
		runCommand = func(name string, stdin string, args ...string) error {
			return errors.New("paste failed")
		}
		p := &windowsPaster{}
		err := p.Paste("hello")
		if err == nil || err.Error() != "failed to paste via PowerShell: paste failed" {
			t.Errorf("expected paste failed error, got: %v", err)
		}
	})
}
