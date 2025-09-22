package licensing

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
	// ErrLicenseNotFound indicates that no license row matches the lookup.
	ErrLicenseNotFound = errors.New("licensing: license not found")
	// ErrFingerprintMismatch is returned when a license key does not belong to the provided fingerprint.
	ErrFingerprintMismatch = errors.New("licensing: fingerprint mismatch")
	// ErrLicenseExpired is returned when the stored license has passed its expiry.
	ErrLicenseExpired = errors.New("licensing: license expired")
)

// LicenseRecord represents a single license stored in the backing database.
type LicenseRecord struct {
	Key             string
	FingerprintHash string
	CustomerID      string
	IssuedAt        time.Time
	ExpiresAt       time.Time
}

// Service coordinates persistence and validation for license keys.
type Service struct {
	db            *sql.DB
	defaultExpiry time.Duration
}

// NewService opens (or creates) the SQLite database used for licensing metadata.
// The defaultExpiry parameter controls how long new licenses remain valid; pass zero
// to indicate licenses should not expire automatically.
func NewService(path string, defaultExpiry time.Duration) (*Service, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("licensing: database path is required")
	}
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("licensing: create database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("licensing: open database: %w", err)
	}

	svc := &Service{db: db, defaultExpiry: defaultExpiry}
	if err := svc.migrate(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	return svc, nil
}

// Close flushes and releases the underlying database resources.
func (s *Service) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// IssueLicense either returns the existing license for the fingerprint or mints
// a new one when none exists or the previous license has expired.
func (s *Service) IssueLicense(ctx context.Context, customerID, fingerprint string) (*LicenseRecord, bool, error) {
	if s == nil {
		return nil, false, errors.New("licensing: service is nil")
	}
	customerID = strings.TrimSpace(customerID)
	fingerprint = strings.TrimSpace(fingerprint)
	if customerID == "" {
		return nil, false, errors.New("licensing: customer identifier is required")
	}
	if fingerprint == "" {
		return nil, false, errors.New("licensing: fingerprint is required")
	}

	hash := HashFingerprint(fingerprint)
	sanitized := sanitizeCustomerID(customerID)
	now := time.Now().UTC()

	existing, err := s.findByFingerprint(ctx, hash)
	if err == nil {
		if !existing.ExpiresAt.IsZero() && !existing.ExpiresAt.After(now) {
			// The previous license expired; issue a replacement in-place.
			return s.updateLicense(ctx, existing, sanitized, hash, now)
		}
		return existing, false, nil
	}
	if !errors.Is(err, ErrLicenseNotFound) {
		return nil, false, err
	}

	key, err := GenerateLicenseKey(sanitized, hash)
	if err != nil {
		return nil, false, err
	}
	record := &LicenseRecord{
		Key:             key,
		FingerprintHash: hash,
		CustomerID:      sanitized,
		IssuedAt:        now,
	}
	if s.defaultExpiry > 0 {
		record.ExpiresAt = now.Add(s.defaultExpiry)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO licenses(key, fingerprint_hash, customer_id, issued_at, expires_at) VALUES(?, ?, ?, ?, ?)`,
		record.Key, record.FingerprintHash, record.CustomerID, record.IssuedAt, nullTime(record.ExpiresAt),
	)
	if err != nil {
		return nil, false, fmt.Errorf("licensing: insert license: %w", err)
	}
	return record, true, nil
}

// ValidateLicense confirms that the provided key belongs to the supplied fingerprint.
func (s *Service) ValidateLicense(ctx context.Context, key, fingerprint string) (*LicenseRecord, error) {
	if s == nil {
		return nil, errors.New("licensing: service is nil")
	}
	key = strings.TrimSpace(key)
	fingerprint = strings.TrimSpace(fingerprint)
	if key == "" {
		return nil, errors.New("licensing: license key is required")
	}
	if fingerprint == "" {
		return nil, errors.New("licensing: fingerprint is required")
	}

	hash := HashFingerprint(fingerprint)
	record, err := s.findByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if record.FingerprintHash != hash {
		return nil, ErrFingerprintMismatch
	}
	if !record.ExpiresAt.IsZero() && record.ExpiresAt.Before(time.Now()) {
		return nil, ErrLicenseExpired
	}
	return record, nil
}

// updateLicense replaces the key for an existing fingerprint when the previous
// license has expired.
func (s *Service) updateLicense(ctx context.Context, existing *LicenseRecord, customerID, fingerprintHash string, issuedAt time.Time) (*LicenseRecord, bool, error) {
	key, err := GenerateLicenseKey(customerID, fingerprintHash)
	if err != nil {
		return nil, false, err
	}
	expiresAt := time.Time{}
	if s.defaultExpiry > 0 {
		expiresAt = issuedAt.Add(s.defaultExpiry)
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE licenses SET key = ?, customer_id = ?, issued_at = ?, expires_at = ? WHERE fingerprint_hash = ?`,
		key, customerID, issuedAt, nullTime(expiresAt), fingerprintHash,
	)
	if err != nil {
		return nil, false, fmt.Errorf("licensing: update license: %w", err)
	}
	existing.Key = key
	existing.CustomerID = customerID
	existing.IssuedAt = issuedAt
	existing.ExpiresAt = expiresAt
	return existing, true, nil
}

func (s *Service) migrate(ctx context.Context) error {
	schema := `
CREATE TABLE IF NOT EXISTS licenses (
    key TEXT PRIMARY KEY,
    fingerprint_hash TEXT NOT NULL UNIQUE,
    customer_id TEXT NOT NULL,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_licenses_fingerprint ON licenses(fingerprint_hash);
`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("licensing: migrate schema: %w", err)
	}
	return nil
}

func (s *Service) findByFingerprint(ctx context.Context, fingerprintHash string) (*LicenseRecord, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT key, fingerprint_hash, customer_id, issued_at, expires_at FROM licenses WHERE fingerprint_hash = ?`,
		fingerprintHash,
	)
	record := &LicenseRecord{}
	var expires sql.NullTime
	if err := row.Scan(&record.Key, &record.FingerprintHash, &record.CustomerID, &record.IssuedAt, &expires); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenseNotFound
		}
		return nil, fmt.Errorf("licensing: query fingerprint: %w", err)
	}
	if expires.Valid {
		record.ExpiresAt = expires.Time
	}
	return record, nil
}

func (s *Service) findByKey(ctx context.Context, key string) (*LicenseRecord, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT key, fingerprint_hash, customer_id, issued_at, expires_at FROM licenses WHERE key = ?`,
		key,
	)
	record := &LicenseRecord{}
	var expires sql.NullTime
	if err := row.Scan(&record.Key, &record.FingerprintHash, &record.CustomerID, &record.IssuedAt, &expires); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLicenseNotFound
		}
		return nil, fmt.Errorf("licensing: query key: %w", err)
	}
	if expires.Valid {
		record.ExpiresAt = expires.Time
	}
	return record, nil
}

func nullTime(value time.Time) interface{} {
	if value.IsZero() {
		return nil
	}
	return value
}
