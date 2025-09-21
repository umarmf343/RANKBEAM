package main

import (
	"strings"
	"testing"
)

func TestSanitizeCustomerID(t *testing.T) {
	got := sanitizeCustomerID(" Acme Corp !123 ")
	if got != "ACMECORP123" {
		t.Fatalf("expected ACMECORP123, got %s", got)
	}
	got = sanitizeCustomerID("!!!")
	if got != "CUSTOMER" {
		t.Fatalf("expected CUSTOMER fallback, got %s", got)
	}
}

func TestGenerateLicenseKey(t *testing.T) {
	hash := HashFingerprint("fingerprint")
	key, err := GenerateLicenseKey("Example", hash)
	if err != nil {
		t.Fatalf("GenerateLicenseKey error: %v", err)
	}
	if len(key) == 0 {
		t.Fatal("expected non-empty key")
	}
	if !strings.HasPrefix(key, "EXAMPLE") {
		t.Fatalf("expected prefix to include sanitized customer id, got %s", key)
	}
}
