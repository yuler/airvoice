//go:build darwin

package paste

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type commandCall struct {
	name  string
	stdin string
	args  []string
}

func TestDarwinPasters(t *testing.T) {
	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()

	var calls []commandCall
	runCommand = func(name string, stdin string, args ...string) error {
		calls = append(calls, commandCall{name: name, stdin: stdin, args: args})
		return nil
	}

	t.Run("darwinPaster success", func(t *testing.T) {
		calls = nil
		p := &darwinPaster{}
		if p.Name() != "darwin" {
			t.Errorf("expected darwin, got %s", p.Name())
		}
		err := p.Paste("hello world")
		if err != nil {
			t.Fatalf("Paste failed: %v", err)
		}
		expected := []commandCall{
			{name: "pbcopy", stdin: "hello world", args: nil},
			{name: "osascript", stdin: "", args: []string{"-e", pasteKeystrokeScript}},
		}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("got calls %+v, expected %+v", calls, expected)
		}
	})

	t.Run("darwinPaster first command failure", func(t *testing.T) {
		runCommand = func(name string, stdin string, args ...string) error {
			if name == "pbcopy" {
				return errors.New("pbcopy failed")
			}
			return nil
		}
		p := &darwinPaster{}
		err := p.Paste("hello")
		if err == nil || !strings.Contains(err.Error(), "pbcopy failed") {
			t.Errorf("expected pbcopy failed error, got: %v", err)
		}
	})

	t.Run("New function switches", func(t *testing.T) {
		origGOOS := goos
		defer func() {
			goos = origGOOS
		}()

		goos = "darwin"
		p, err := New()
		if err != nil {
			t.Fatalf("New failed on darwin: %v", err)
		}
		if p.Name() != "darwin" {
			t.Errorf("expected darwin, got %s", p.Name())
		}

		goos = "windows"
		_, err = New()
		if err == nil {
			t.Error("expected error on windows session, got nil")
		}
	})

	t.Run("formatDarwinKeystrokeError accessibility", func(t *testing.T) {
		err := formatDarwinKeystrokeError(fmt.Errorf(`36:68: execution error: System Events got an error: osascript is not allowed to send keystrokes. (1002)`))
		if !strings.Contains(err.Error(), "辅助功能") {
			t.Fatalf("expected accessibility hint, got: %v", err)
		}
	})
}
