package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestHandlePaystackWebhookCreatesLicense(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")
	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	secret := "paystack-secret"
	handler := NewLicenseHandler(store, "", secret)

	payload := map[string]any{
		"event": "charge.success",
		"data": map[string]any{
			"reference": "PSK_ref_123",
			"paid_at":   time.Now().UTC().Format(time.RFC3339),
			"customer": map[string]any{
				"email": "user@example.com",
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/paystack/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-paystack-signature", signPayload(secret, body))
	rr := httptest.NewRecorder()

	handler.HandlePaystackWebhook(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	var resp struct {
		LicenseKey string `json:"licenseKey"`
		ExpiresAt  string `json:"expiresAt"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.LicenseKey == "" {
		t.Fatal("expected license key in response")
	}
	if resp.ExpiresAt == "" {
		t.Fatal("expected expiry timestamp in response")
	}

	if _, err := store.ValidateLicense(context.Background(), resp.LicenseKey, "user@example.com"); err != nil {
		t.Fatalf("ValidateLicense: %v", err)
	}
}

func TestValidateLicenseRequiresToken(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")
	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	handler := NewLicenseHandler(store, "installer-secret", "")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/licenses/validate", bytes.NewReader([]byte(`{"licenseKey":"key","email":"user@example.com"}`)))
	rr := httptest.NewRecorder()

	handler.ValidateLicense(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestValidateLicenseSuccess(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")
	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	expiresAt := time.Now().Add(licenseValidity)
	lic, err := store.CreateLicense(context.Background(), "user@example.com", "ref-123", expiresAt)
	if err != nil {
		t.Fatalf("CreateLicense: %v", err)
	}

	handler := NewLicenseHandler(store, "installer-secret", "")

	body, _ := json.Marshal(map[string]string{
		"licenseKey": lic.Key,
		"email":      "user@example.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/licenses/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Installer-Token", "installer-secret")
	rr := httptest.NewRecorder()

	handler.ValidateLicense(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestHandlePaystackWebhookInvalidSignature(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")
	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	handler := NewLicenseHandler(store, "", "paystack-secret")

	payload := map[string]any{"event": "charge.success", "data": map[string]any{"reference": "ref", "customer": map[string]any{"email": "user@example.com"}}}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/paystack/webhook", bytes.NewReader(body))
	req.Header.Set("x-paystack-signature", "invalid")
	rr := httptest.NewRecorder()

	handler.HandlePaystackWebhook(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func signPayload(secret string, body []byte) string {
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
