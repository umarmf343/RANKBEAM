package license

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSaveAndLoadLicenseKey(t *testing.T) {
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
	}

	savedPath, err := SaveLicenseKey("TEST-KEY-123")
	if err != nil {
		t.Fatalf("SaveLicenseKey error: %v", err)
	}
	if !strings.HasPrefix(savedPath, dir) {
		t.Fatalf("expected storage path under %s, got %s", dir, savedPath)
	}

	key, err := LoadLicenseKey()
	if err != nil {
		t.Fatalf("LoadLicenseKey error: %v", err)
	}
	if key != "TEST-KEY-123" {
		t.Fatalf("expected key TEST-KEY-123, got %s", key)
	}
}

func TestLoadLicenseKeyEmpty(t *testing.T) {
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
	}

	path, err := StoragePath()
	if err != nil {
		t.Fatalf("StoragePath error: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("\n\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err = LoadLicenseKey()
	if !errors.Is(err, ErrEmptyLicenseKey) {
		t.Fatalf("expected ErrEmptyLicenseKey, got %v", err)
	}
}
