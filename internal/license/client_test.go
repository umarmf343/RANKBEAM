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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", srv.Client())
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	if err := client.ValidateLicense(context.Background(), "KEY", "FP"); err != nil {
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

	err = client.ValidateLicense(context.Background(), "KEY", "FP")
	if err != ErrInvalidLicense {
		t.Fatalf("expected ErrInvalidLicense, got %v", err)
	}
}

func TestRequestLicense(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/licenses" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"licenseKey": "KEY"})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", srv.Client())
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	key, err := client.RequestLicense(context.Background(), "customer", "fingerprint")
	if err != nil {
		t.Fatalf("RequestLicense error: %v", err)
	}
	if key != "KEY" {
		t.Fatalf("expected KEY, got %s", key)
	}
}

func TestValidateLocalLicense(t *testing.T) {
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
	}

	if _, err := SaveLicenseKey("LOCAL-KEY"); err != nil {
		t.Fatalf("SaveLicenseKey: %v", err)
	}
	fingerprint, err := Fingerprint()
	if err != nil {
		t.Fatalf("Fingerprint error: %v", err)
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
		if payload["fingerprint"] != fingerprint {
			t.Fatalf("unexpected fingerprint %s", payload["fingerprint"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "", srv.Client())
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	key, err := ValidateLocalLicense(context.Background(), client)
	if err != nil {
		t.Fatalf("ValidateLocalLicense error: %v", err)
	}
	if key != "LOCAL-KEY" {
		t.Fatalf("expected LOCAL-KEY, got %s", key)
	}
}
