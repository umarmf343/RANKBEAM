package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestLicenseStoreLifecycle(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")

	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	expiresAt := time.Now().Add(licenseValidity)

	lic, created, err := store.CreateLicense(ctx, "user@example.com", "ref-123", expiresAt)
	if err != nil {
		t.Fatalf("CreateLicense: %v", err)
	}
	if !created {
		t.Fatalf("expected license to be created")
	}
	if lic.Key == "" {
		t.Fatalf("expected license key to be set")
	}

	fetched, err := store.FindByKey(ctx, lic.Key)
	if err != nil {
		t.Fatalf("FindByKey: %v", err)
	}
	if fetched.CustomerEmail != "user@example.com" {
		t.Fatalf("expected stored email, got %s", fetched.CustomerEmail)
	}

	if _, err := store.ValidateLicense(ctx, lic.Key, "user@example.com"); err != nil {
		t.Fatalf("ValidateLicense: %v", err)
	}

	if _, err := store.ValidateLicense(ctx, lic.Key, "wrong@example.com"); !errors.Is(err, ErrEmailMismatch) {
		t.Fatalf("expected ErrEmailMismatch, got %v", err)
	}
}

func TestLicenseExpiration(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")

	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	expiresAt := time.Now().Add(-time.Hour)
	lic, created, err := store.CreateLicense(ctx, "user@example.com", "ref-123", expiresAt)
	if err != nil {
		t.Fatalf("CreateLicense: %v", err)
	}
	if !created {
		t.Fatalf("expected license to be created")
	}

	if _, err := store.ValidateLicense(ctx, lic.Key, "user@example.com"); !errors.Is(err, ErrLicenseExpired) {
		t.Fatalf("expected ErrLicenseExpired, got %v", err)
	}
}

func TestCreateLicenseIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")

	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	expiresAt := time.Now().Add(licenseValidity)

	lic1, created1, err := store.CreateLicense(ctx, "user@example.com", "ref-123", expiresAt)
	if err != nil {
		t.Fatalf("CreateLicense: %v", err)
	}
	if !created1 {
		t.Fatalf("expected first call to create a record")
	}

	lic2, created2, err := store.CreateLicense(ctx, "user@example.com", "ref-123", expiresAt)
	if err != nil {
		t.Fatalf("CreateLicense: %v", err)
	}
	if created2 {
		t.Fatalf("expected second call to reuse existing record")
	}
	if lic1.Key != lic2.Key {
		t.Fatalf("expected same license key, got %s vs %s", lic1.Key, lic2.Key)
	}
	if !lic1.CreatedAt.Equal(lic2.CreatedAt) {
		t.Fatalf("expected created timestamp to match")
	}
}
