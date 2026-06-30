package main

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type HistoryEntry struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Device    string    `json:"device"`
	CreatedAt time.Time `json:"createdAt"`
}

type HistoryStore struct {
	db *sql.DB
}

func NewHistoryStore(dbPath string) (*HistoryStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		db.Close()
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			content TEXT NOT NULL,
			device TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &HistoryStore{db: db}, nil
}

func (h *HistoryStore) Add(content, device string) error {
	_, err := h.db.Exec(
		"INSERT INTO history (content, device) VALUES (?, ?)",
		content, device,
	)
	return err
}

func (h *HistoryStore) List(limit int) ([]HistoryEntry, error) {
	rows, err := h.db.Query(
		"SELECT id, content, device, created_at FROM history ORDER BY id DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []HistoryEntry{}
	for rows.Next() {
		var e HistoryEntry
		if err := rows.Scan(&e.ID, &e.Content, &e.Device, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (h *HistoryStore) Clear() error {
	_, err := h.db.Exec("DELETE FROM history")
	return err
}

func (h *HistoryStore) Search(query string, limit int) ([]HistoryEntry, error) {
	rows, err := h.db.Query(
		"SELECT id, content, device, created_at FROM history WHERE content LIKE ? ORDER BY id DESC LIMIT ?",
		"%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []HistoryEntry{}
	for rows.Next() {
		var e HistoryEntry
		if err := rows.Scan(&e.ID, &e.Content, &e.Device, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (h *HistoryStore) Close() error {
	return h.db.Close()
}
