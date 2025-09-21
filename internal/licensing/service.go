package licensing

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Service coordinates key generation, persistence and validation.
type Service struct {
	store         *Store
	defaultExpiry time.Duration
}

// License captures the persistent representation of an issued key.
type License struct {
	Key         string
	Fingerprint string
	CustomerID  string
	IssuedAt    time.Time
	ExpiresAt   *time.Time
}

// NewService constructs a new Service backed by the SQLite database at path.
func NewService(dbPath string, defaultExpiry time.Duration) (*Service, error) {
	store, err := OpenStore(dbPath)
	if err != nil {
		return nil, err
	}
	return &Service{store: store, defaultExpiry: defaultExpiry}, nil
}

// Close releases any underlying store resources.
func (s *Service) Close() error {
	if s == nil {
		return nil
	}
	return s.store.Close()
}

// IssueLicense generates or reuses a license for the provided fingerprint.
func (s *Service) IssueLicense(ctx context.Context, customerID, fingerprint string) (License, error) {
	if customerID == "" {
		return License{}, errors.New("customer identifier is required")
	}
	if fingerprint == "" {
		return License{}, errors.New("fingerprint is required")
	}

	var existing License
	row := s.store.QueryRow(ctx, "SELECT license_key, fingerprint, customer_id, issued_at, expires_at FROM licenses WHERE fingerprint = ?", fingerprint)
	switch err := row.Scan(&existing.Key, &existing.Fingerprint, &existing.CustomerID, &existing.IssuedAt, &existing.ExpiresAt); {
	case err == nil:
		if existing.CustomerID != customerID {
			return License{}, fmt.Errorf("fingerprint already registered to customer %s", existing.CustomerID)
		}
		if existing.ExpiresAt != nil && existing.ExpiresAt.Before(time.Now()) {
			// Renew expired license.
			return s.refreshLicense(ctx, existing.Key, customerID, fingerprint)
		}
		return existing, nil
	case errors.Is(err, sql.ErrNoRows):
		// Proceed to generate a new license.
	default:
		return License{}, err
	}

	key, err := generateKey(customerID, fingerprint)
	if err != nil {
		return License{}, err
	}

	issuedAt := time.Now().UTC()
	var expiresAt *time.Time
	if s.defaultExpiry > 0 {
		expiry := issuedAt.Add(s.defaultExpiry)
		expiresAt = &expiry
	}

	var expiresValue any
	if expiresAt != nil {
		expiresValue = *expiresAt
	}

	if err := s.store.Exec(ctx, "INSERT INTO licenses (license_key, fingerprint, customer_id, issued_at, expires_at) VALUES (?, ?, ?, ?, ?)", key, fingerprint, customerID, issuedAt, expiresValue); err != nil {
		return License{}, err
	}

	return License{Key: key, Fingerprint: fingerprint, CustomerID: customerID, IssuedAt: issuedAt, ExpiresAt: expiresAt}, nil
}

// refreshLicense updates the issue/expiry timestamps for an existing key.
func (s *Service) refreshLicense(ctx context.Context, key, customerID, fingerprint string) (License, error) {
	issuedAt := time.Now().UTC()
	var expiresAt *time.Time
	if s.defaultExpiry > 0 {
		expiry := issuedAt.Add(s.defaultExpiry)
		expiresAt = &expiry
	}

	var expiresValue any
	if expiresAt != nil {
		expiresValue = *expiresAt
	}

	if err := s.store.Exec(ctx, "UPDATE licenses SET issued_at = ?, expires_at = ? WHERE license_key = ?", issuedAt, expiresValue, key); err != nil {
		return License{}, err
	}
	return License{Key: key, Fingerprint: fingerprint, CustomerID: customerID, IssuedAt: issuedAt, ExpiresAt: expiresAt}, nil
}

// ValidateLicense ensures the key exists, matches the supplied fingerprint and is not expired.
func (s *Service) ValidateLicense(ctx context.Context, key, fingerprint string) (License, error) {
	if key == "" || fingerprint == "" {
		return License{}, errors.New("license key and fingerprint are required")
	}

	var lic License
	row := s.store.QueryRow(ctx, "SELECT license_key, fingerprint, customer_id, issued_at, expires_at FROM licenses WHERE license_key = ?", key)
	if err := row.Scan(&lic.Key, &lic.Fingerprint, &lic.CustomerID, &lic.IssuedAt, &lic.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return License{}, errors.New("license not found")
		}
		return License{}, err
	}

	if !strings.EqualFold(lic.Fingerprint, fingerprint) {
		return License{}, errors.New("fingerprint mismatch")
	}

	if lic.ExpiresAt != nil && time.Now().After(*lic.ExpiresAt) {
		return License{}, errors.New("license expired")
	}

	return lic, nil
}

func generateKey(customerID, fingerprint string) (string, error) {
	randomSeed := make([]byte, 32)
	if _, err := rand.Read(randomSeed); err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, randomSeed)
	mac.Write([]byte(strings.ToUpper(customerID)))
	mac.Write([]byte("|"))
	mac.Write([]byte(fingerprint))

	raw := mac.Sum(nil)
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw)

	// Format as groups of 5 characters for readability.
	var groups []string
	for i := 0; i < len(encoded); i += 5 {
		end := i + 5
		if end > len(encoded) {
			end = len(encoded)
		}
		groups = append(groups, encoded[i:end])
	}

	return strings.Join(groups, "-"), nil
}
