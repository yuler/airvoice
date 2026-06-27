package main

import (
	"os"
	"testing"
)

func TestHistoryStore(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "history-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	store, err := NewHistoryStore(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// Test Add
	if err := store.Add("Hello", "iPhone"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Test List
	entries, err := store.List(10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("List() got %d entries, want 1", len(entries))
	}
	if entries[0].Content != "Hello" {
		t.Errorf("Content = %s, want Hello", entries[0].Content)
	}

	// Test Clear
	if err := store.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	entries, _ = store.List(10)
	if len(entries) != 0 {
		t.Errorf("After Clear(), got %d entries, want 0", len(entries))
	}
}
