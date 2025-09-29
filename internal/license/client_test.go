package license

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"
)

func TestValidateLicense(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/paystack/validate" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if token := r.Header.Get("X-License-Token"); token != defaultInstallerToken {
			t.Fatalf("expected X-License-Token %s, got %s", defaultInstallerToken, token)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["email"] != "user@example.com" {
			t.Fatalf("unexpected email %s", payload["email"])
		}
		if payload["licenseKey"] != "KEY" {
			t.Fatalf("unexpected license key %s", payload["licenseKey"])
		}
		if payload["fingerprint"] != "ABC123" {
			t.Fatalf("unexpected fingerprint %s", payload["fingerprint"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "valid", "expiresAt": time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339)})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", srv.Client())
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	expiresAt, err := client.ValidateLicense(context.Background(), "KEY", "user@example.com", "ABC123")
	if err != nil {
		t.Fatalf("ValidateLicense unexpected error: %v", err)
	}
	if expiresAt.IsZero() {
		t.Fatalf("expected non-zero expiry")
	}
}

func TestValidateLicenseInvalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", srv.Client())
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	_, err = client.ValidateLicense(context.Background(), "KEY", "user@example.com", "ABC123")
	if err != ErrInvalidLicense {
		t.Fatalf("expected ErrInvalidLicense, got %v", err)
	}
}

func TestValidateLocalLicense(t *testing.T) {
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
	}

	fingerprint := "LOCALFP"
	if _, err := SaveLicense(LicenseData{Key: "LOCAL-KEY", Email: "User@example.com", Fingerprint: fingerprint}); err != nil {
		t.Fatalf("SaveLicense: %v", err)
	}

	expiry := time.Now().Add(30 * 24 * time.Hour).Truncate(time.Second)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/paystack/validate" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if token := r.Header.Get("X-License-Token"); token != defaultInstallerToken {
			t.Fatalf("expected X-License-Token %s, got %s", defaultInstallerToken, token)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["licenseKey"] != "LOCAL-KEY" {
			t.Fatalf("unexpected license key %s", payload["licenseKey"])
		}
		if payload["email"] != "user@example.com" {
			t.Fatalf("unexpected email %s", payload["email"])
		}
		if payload["fingerprint"] != fingerprint {
			t.Fatalf("unexpected fingerprint %s", payload["fingerprint"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "valid", "expiresAt": expiry.Format(time.RFC3339)})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", srv.Client())
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	data, err := ValidateLocalLicense(context.Background(), client)
	if err != nil {
		t.Fatalf("ValidateLocalLicense error: %v", err)
	}
	if data.Key != "LOCAL-KEY" || data.Email != "user@example.com" || data.Fingerprint != fingerprint {
		t.Fatalf("unexpected license data: %+v", data)
	}
	if data.ExpiresAt.IsZero() {
		t.Fatalf("expected expiry to be set")
	}
}
