package license

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func setConfigEnv(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
	}
}

func TestSaveAndLoadLicense(t *testing.T) {
	dir := t.TempDir()
	setConfigEnv(t, dir)

	expiry := time.Now().Add(12 * time.Hour).UTC().Truncate(time.Second)
	input := LicenseData{Key: "test-key-123", Email: "USER@example.COM", Fingerprint: "DEVICE123", ExpiresAt: expiry}
	savedPath, err := SaveLicense(input)
	if err != nil {
		t.Fatalf("SaveLicense error: %v", err)
	}
	if !strings.HasPrefix(savedPath, dir) {
		t.Fatalf("expected storage path under %s, got %s", dir, savedPath)
	}

	data, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("read saved license: %v", err)
	}
	var stored LicenseData
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("unmarshal saved license: %v", err)
	}
	if stored.Email != "user@example.com" {
		t.Fatalf("expected stored email to be normalised, got %s", stored.Email)
	}

	loaded, err := LoadLicense()
	if err != nil {
		t.Fatalf("LoadLicense error: %v", err)
	}
	if loaded.Key != "TEST-KEY-123" || loaded.Email != "user@example.com" || loaded.Fingerprint != "DEVICE123" {
		t.Fatalf("unexpected loaded data: %+v", loaded)
	}
	if loaded.ExpiresAt.IsZero() {
		t.Fatalf("expected expiry to be set")
	}
}

func TestLoadLicenseEmptyFile(t *testing.T) {
	dir := t.TempDir()
	setConfigEnv(t, dir)

	path, err := StoragePath()
	if err != nil {
		t.Fatalf("StoragePath error: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
		t.Fatalf("write empty file: %v", err)
	}
	_, err = LoadLicense()
	if !errors.Is(err, ErrEmptyLicenseFile) {
		t.Fatalf("expected ErrEmptyLicenseFile, got %v", err)
	}
}

func TestLoadLicenseInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	setConfigEnv(t, dir)

	path, err := StoragePath()
	if err != nil {
		t.Fatalf("StoragePath error: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err = LoadLicense()
	if err == nil || !strings.Contains(err.Error(), "parse stored license") {
		t.Fatalf("expected JSON parse error, got %v", err)
	}
}

func TestLoadLicenseMissingFields(t *testing.T) {
	dir := t.TempDir()
	setConfigEnv(t, dir)

	path, err := StoragePath()
	if err != nil {
		t.Fatalf("StoragePath error: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	encoded, err := json.Marshal(LicenseData{Key: "", Email: ""})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err = LoadLicense()
	if !errors.Is(err, ErrEmptyLicenseKey) {
		t.Fatalf("expected ErrEmptyLicenseKey, got %v", err)
	}

	encoded, err = json.Marshal(LicenseData{Key: "KEY", Email: "user@example.com"})
	if err != nil {
		t.Fatalf("marshal fingerprint test: %v", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err = LoadLicense()
	if !errors.Is(err, ErrMissingFingerprint) {
		t.Fatalf("expected ErrMissingFingerprint, got %v", err)
	}
}
