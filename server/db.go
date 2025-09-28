package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrLicenseNotFound = errors.New("license: not found")
	ErrEmailMismatch   = errors.New("license: email mismatch")
	ErrLicenseExpired  = errors.New("license: expired")
)

const licenseValidity = 30 * 24 * time.Hour

type License struct {
	Key            string
	CustomerEmail  string
	TransactionRef string
	ExpiresAt      time.Time
	CreatedAt      time.Time
}

type LicenseStore struct {
	db *sql.DB
}

func NewLicenseStore(path string) (*LicenseStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	store := &LicenseStore{db: db}
	if err := store.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *LicenseStore) Close() error {
	return s.db.Close()
}

func (s *LicenseStore) migrate(ctx context.Context) error {
	schema := `
CREATE TABLE IF NOT EXISTS licenses (
    key TEXT PRIMARY KEY,
    customer_email TEXT NOT NULL,
    transaction_ref TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_licenses_email ON licenses(customer_email);
`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	return nil
}

func (s *LicenseStore) CreateLicense(ctx context.Context, email, reference string, expiresAt time.Time) (*License, error) {
	key, err := GenerateLicenseKey(email, reference, expiresAt)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	email = strings.TrimSpace(strings.ToLower(email))
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO licenses(key, customer_email, transaction_ref, expires_at, created_at) VALUES(?, ?, ?, ?, ?)`,
		key, email, strings.TrimSpace(reference), expiresAt.UTC(), now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert license: %w", err)
	}

	return &License{Key: key, CustomerEmail: email, TransactionRef: reference, ExpiresAt: expiresAt.UTC(), CreatedAt: now}, nil
}

func (s *LicenseStore) FindByKey(ctx context.Context, key string) (*License, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT key, customer_email, transaction_ref, expires_at, created_at FROM licenses WHERE key = ?`, key,
	)
	lic := &License{}
	if err := row.Scan(&lic.Key, &lic.CustomerEmail, &lic.TransactionRef, &lic.ExpiresAt, &lic.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenseNotFound
		}
		return nil, err
	}
	return lic, nil
}

func (s *LicenseStore) ValidateLicense(ctx context.Context, key, email string) (*License, error) {
	lic, err := s.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(strings.ToLower(email)) == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !emailsEqual(lic.CustomerEmail, email) {
		return nil, ErrEmailMismatch
	}
	if time.Now().UTC().After(lic.ExpiresAt) {
		return nil, ErrLicenseExpired
	}
	return lic, nil
}

func emailsEqual(stored, provided string) bool {
	return strings.EqualFold(strings.TrimSpace(stored), strings.TrimSpace(provided))
}
