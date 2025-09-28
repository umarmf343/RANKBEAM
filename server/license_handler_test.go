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

type fakeMailer struct {
        calls   int
        lastTo  string
        lastKey string
}

func (f *fakeMailer) SendLicenseEmail(ctx context.Context, to string, license *License) error {
        f.calls++
        f.lastTo = to
        if license != nil {
                f.lastKey = license.Key
        }
        return nil
}

func TestHandlePaystackWebhookCreatesLicense(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")
	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	secret := "paystack-secret"
        mailer := &fakeMailer{}
        handler := NewLicenseHandler(store, "", secret, mailer)

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

        if mailer.calls != 1 {
                t.Fatalf("expected mailer to be called once, got %d", mailer.calls)
        }
        if mailer.lastTo != "user@example.com" {
                t.Fatalf("expected email to be sent to user@example.com, got %s", mailer.lastTo)
        }
        if mailer.lastKey != resp.LicenseKey {
                t.Fatalf("expected email to include license key %s, got %s", resp.LicenseKey, mailer.lastKey)
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

        handler := NewLicenseHandler(store, "installer-secret", "", nil)

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

        handler := NewLicenseHandler(store, "installer-secret", "", nil)

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

        handler := NewLicenseHandler(store, "", "paystack-secret", nil)

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
