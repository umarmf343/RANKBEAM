package license

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	productConfigDir = "RankBeam"
	licenseFileName  = "license.json"
)

var ErrEmptyLicenseFile = errors.New("license: empty license file")

// StoragePath returns the absolute path where the installer stores the
// machine's license data.
func StoragePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("license: resolve config dir: %w", err)
	}
	return filepath.Join(configDir, productConfigDir, licenseFileName), nil
}

// LoadLicense reads and normalises the persisted license details.
func LoadLicense() (LicenseData, error) {
	path, err := StoragePath()
	if err != nil {
		return LicenseData{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return LicenseData{}, err
	}
	if len(data) == 0 {
		return LicenseData{}, ErrEmptyLicenseFile
	}
	var payload LicenseData
	if err := json.Unmarshal(data, &payload); err != nil {
		return LicenseData{}, fmt.Errorf("license: parse stored license: %w", err)
	}
	payload = payload.Normalise()
	if err := payload.Validate(); err != nil {
		return LicenseData{}, err
	}
	return payload, nil
}

// SaveLicense writes the license data to the storage path. The file is stored as JSON
// to capture both the key and associated email address.
func SaveLicense(data LicenseData) (string, error) {
	normalised := data.Normalise()
	if err := normalised.Validate(); err != nil {
		return "", err
	}
	path, err := StoragePath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("license: create config dir: %w", err)
	}
	encoded, err := json.MarshalIndent(normalised, "", "  ")
	if err != nil {
		return "", fmt.Errorf("license: encode license: %w", err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("license: write license: %w", err)
	}
	return path, nil
}
