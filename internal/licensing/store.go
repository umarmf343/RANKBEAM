package licensing

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `CREATE TABLE IF NOT EXISTS licenses (
    license_key TEXT PRIMARY KEY,
    fingerprint TEXT NOT NULL,
    customer_id TEXT NOT NULL,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP
);`

// Store wraps access to the SQLite-backed license registry.
type Store struct {
	db *sql.DB
}

// OpenStore opens (or creates) the SQLite database located at path. The schema is
// initialized on first use.
func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

// Close releases underlying database resources.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Exec executes a statement with the provided arguments.
func (s *Store) Exec(ctx context.Context, query string, args ...any) error {
	if s == nil || s.db == nil {
		return errors.New("store is not initialized")
	}
	_, err := s.db.ExecContext(ctx, query, args...)
	return err
}

// QueryRow runs a row query against the store.
func (s *Store) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

// Query executes a query that returns multiple rows.
func (s *Store) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}
