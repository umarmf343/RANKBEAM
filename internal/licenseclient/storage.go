package licenseclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Storage persists the license payload to disk.
type Storage struct {
	path string
}

// NewStorage resolves the configuration directory and returns a storage handle.
func NewStorage(appID string) (*Storage, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("locate config dir: %w", err)
	}

	if appID == "" {
		appID = "amazon-product-suite"
	}

	path := filepath.Join(dir, appID, "license.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	return &Storage{path: path}, nil
}

// Path returns the resolved file path used for persistence.
func (s *Storage) Path() string {
	return s.path
}

// Save writes the envelope to disk.
func (s *Storage) Save(envelope LicenseEnvelope) error {
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

// Load retrieves the envelope from disk.
func (s *Storage) Load() (LicenseEnvelope, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return LicenseEnvelope{}, os.ErrNotExist
	}
	if err != nil {
		return LicenseEnvelope{}, err
	}

	var envelope LicenseEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return LicenseEnvelope{}, err
	}
	return envelope, nil
}
