package license

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

func TestValidateLicense(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/licenses/validate" {
			t.Fatalf("unexpected path %s", r.URL.Path)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", srv.Client())
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	if err := client.ValidateLicense(context.Background(), "KEY", "user@example.com"); err != nil {
		t.Fatalf("ValidateLicense unexpected error: %v", err)
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

	err = client.ValidateLicense(context.Background(), "KEY", "user@example.com")
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

	if _, err := SaveLicense(LicenseData{Key: "LOCAL-KEY", Email: "User@example.com"}); err != nil {
		t.Fatalf("SaveLicense: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/licenses/validate" {
			t.Fatalf("unexpected path %s", r.URL.Path)
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
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
	if data.Key != "LOCAL-KEY" || data.Email != "user@example.com" {
		t.Fatalf("unexpected license data: %+v", data)
	}
}
