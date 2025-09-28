package main

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateLicenseKey(t *testing.T) {
	expiresAt := time.Date(2025, time.October, 1, 0, 0, 0, 0, time.UTC)
	key, err := GenerateLicenseKey("user@example.com", "PSK_ref", expiresAt)
	if err != nil {
		t.Fatalf("GenerateLicenseKey error: %v", err)
	}
	if key == "" {
		t.Fatal("expected non-empty key")
	}
	if !strings.HasSuffix(key, "-2025-10-01") {
		t.Fatalf("expected expiry suffix, got %s", key)
	}

	parts := strings.SplitN(key, "-", 2)
	if len(parts[0]) <= 10 {
		t.Fatalf("expected combined hash prefix, got %s", key)
	}
}

func TestGenerateLicenseKeyRequiresFields(t *testing.T) {
	_, err := GenerateLicenseKey("", "ref", time.Now())
	if err == nil {
		t.Fatal("expected error for missing email")
	}
	_, err = GenerateLicenseKey("user@example.com", "", time.Now())
	if err == nil {
		t.Fatal("expected error for missing reference")
	}
}
