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
	ErrLicenseNotFound     = errors.New("license: not found")
	ErrFingerprintMismatch = errors.New("license: fingerprint mismatch")
)

type License struct {
	Key             string
	FingerprintHash string
	CustomerID      string
	CreatedAt       time.Time
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
    fingerprint_hash TEXT NOT NULL UNIQUE,
    customer_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate schema: %w", err)
	}
	return nil
}

func (s *LicenseStore) CreateLicense(ctx context.Context, customerID, fingerprintHash string) (*License, bool, error) {
	if existing, err := s.FindByFingerprint(ctx, fingerprintHash); err == nil {
		return existing, false, nil
	} else if !errors.Is(err, ErrLicenseNotFound) {
		return nil, false, err
	}

	key, err := GenerateLicenseKey(customerID, fingerprintHash)
	if err != nil {
		return nil, false, err
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO licenses(key, fingerprint_hash, customer_id, created_at) VALUES(?, ?, ?, ?)`,
		key, fingerprintHash, customerID, now,
	)
	if err != nil {
		if sqliteIsConstraint(err) {
			if existing, findErr := s.FindByFingerprint(ctx, fingerprintHash); findErr == nil {
				return existing, false, nil
			}
		}
		return nil, false, fmt.Errorf("insert license: %w", err)
	}

	return &License{Key: key, FingerprintHash: fingerprintHash, CustomerID: customerID, CreatedAt: now}, true, nil
}

func (s *LicenseStore) FindByFingerprint(ctx context.Context, fingerprintHash string) (*License, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT key, fingerprint_hash, customer_id, created_at FROM licenses WHERE fingerprint_hash = ?`, fingerprintHash,
	)
	lic := &License{}
	if err := row.Scan(&lic.Key, &lic.FingerprintHash, &lic.CustomerID, &lic.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenseNotFound
		}
		return nil, err
	}
	return lic, nil
}

func (s *LicenseStore) FindByKey(ctx context.Context, key string) (*License, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT key, fingerprint_hash, customer_id, created_at FROM licenses WHERE key = ?`, key,
	)
	lic := &License{}
	if err := row.Scan(&lic.Key, &lic.FingerprintHash, &lic.CustomerID, &lic.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenseNotFound
		}
		return nil, err
	}
	return lic, nil
}

func (s *LicenseStore) ValidateLicense(ctx context.Context, key, fingerprintHash string) (*License, error) {
	lic, err := s.FindByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if lic.FingerprintHash != fingerprintHash {
		return nil, ErrFingerprintMismatch
	}
	return lic, nil
}

func sqliteIsConstraint(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "constraint")
}
