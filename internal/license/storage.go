package license

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const productConfigDir = "AmazonProductSuite"

var ErrEmptyLicenseKey = errors.New("license: empty key on disk")

// StoragePath returns the absolute path where the installer stores the
// machine's license key.
func StoragePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("license: resolve config dir: %w", err)
	}
	return filepath.Join(configDir, productConfigDir, "license.key"), nil
}

// LoadLicenseKey reads and trims the persisted license key.
func LoadLicenseKey() (string, error) {
	path, err := StoragePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", ErrEmptyLicenseKey
	}
	return key, nil
}

// SaveLicenseKey writes the license key to the storage path. It is primarily
// used by tests and automation. The installer normally performs this step.
func SaveLicenseKey(key string) (string, error) {
	path, err := StoragePath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("license: create config dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimSpace(key)+"\n"), 0o600); err != nil {
		return "", fmt.Errorf("license: write key: %w", err)
	}
	return path, nil
}
