package storage

import (
	"database/sql"
	"errors"

	"github.com/AnkushinDaniil/memex/internal/memory"
	_ "github.com/mattn/go-sqlite3"
)

// ErrNotImplemented is returned when a method is not yet implemented
var ErrNotImplemented = errors.New("not yet implemented")

// Ensure SQLiteStorage implements memory.Storage
var _ memory.Storage = (*SQLiteStorage)(nil)

// SQLiteStorage implements Storage using SQLite with FTS5
type SQLiteStorage struct {
	db   *sql.DB
	path string
}

// NewSQLite creates a new SQLite storage instance
func NewSQLite(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return &SQLiteStorage{
		db:   db,
		path: path,
	}, nil
}

// Initialize creates the database schema (stub)
func (s *SQLiteStorage) Initialize() error {
	// Stub - will be implemented in subtask 03
	return nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// Create stores a new memory (stub)
func (s *SQLiteStorage) Create(mem *memory.Memory) error {
	// Stub - will be implemented in subtask 03
	return nil
}

// Get retrieves a memory by ID (stub)
func (s *SQLiteStorage) Get(id string) (*memory.Memory, error) {
	// Stub - will be implemented in subtask 03
	return nil, ErrNotImplemented
}

// Delete removes a memory (stub)
func (s *SQLiteStorage) Delete(id string) error {
	// Stub - will be implemented in subtask 04
	return nil
}

// Search performs full-text search (stub)
func (s *SQLiteStorage) Search(query memory.SearchQuery) ([]memory.SearchResult, error) {
	// Stub - will be implemented in subtask 04
	return nil, ErrNotImplemented
}

// List returns recent memories (stub)
func (s *SQLiteStorage) List(projectID string, limit int, tags []string) ([]memory.Memory, error) {
	// Stub - will be implemented in subtask 04
	return nil, ErrNotImplemented
}
