package main

import (
	"context"
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
	hash := HashFingerprint("fingerprint")

	lic, created, err := store.CreateLicense(ctx, "CUSTOMER", hash)
	if err != nil {
		t.Fatalf("CreateLicense: %v", err)
	}
	if !created {
		t.Fatalf("expected license to be newly created")
	}

	lic2, created, err := store.CreateLicense(ctx, "CUSTOMER", hash)
	if err != nil {
		t.Fatalf("CreateLicense second call: %v", err)
	}
	if created {
		t.Fatalf("expected existing license to be returned")
	}
	if lic2.Key != lic.Key {
		t.Fatalf("expected same key, got %s vs %s", lic2.Key, lic.Key)
	}

	validated, err := store.ValidateLicense(ctx, lic.Key, hash)
	if err != nil {
		t.Fatalf("ValidateLicense: %v", err)
	}
	if validated.Key != lic.Key {
		t.Fatalf("unexpected validated key %s", validated.Key)
	}

	if _, err := store.ValidateLicense(ctx, lic.Key, HashFingerprint("other")); err != ErrFingerprintMismatch {
		t.Fatalf("expected ErrFingerprintMismatch, got %v", err)
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
	hash := HashFingerprint("fingerprint")

	lic, created, err := store.CreateLicense(ctx, "CUSTOMER", hash)
	if err != nil {
		t.Fatalf("CreateLicense: %v", err)
	}
	if !created {
		t.Fatalf("expected license to be newly created")
	}

	expiredAt := time.Now().UTC().Add(-31 * 24 * time.Hour)
	if _, err := store.db.ExecContext(ctx,
		`UPDATE licenses SET created_at = ? WHERE key = ?`,
		expiredAt, lic.Key,
	); err != nil {
		t.Fatalf("update license timestamp: %v", err)
	}

	if _, err := store.ValidateLicense(ctx, lic.Key, hash); err != ErrLicenseExpired {
		t.Fatalf("expected ErrLicenseExpired, got %v", err)
	}

	refreshed, created, err := store.CreateLicense(ctx, "CUSTOMER", hash)
	if err != nil {
		t.Fatalf("CreateLicense after expiry: %v", err)
	}
	if !created {
		t.Fatalf("expected new license to be issued after expiry")
	}
	if refreshed.Key == lic.Key {
		t.Fatalf("expected new license key after expiry")
	}

	if _, err := store.ValidateLicense(ctx, refreshed.Key, hash); err != nil {
		t.Fatalf("ValidateLicense after refresh: %v", err)
	}
}
