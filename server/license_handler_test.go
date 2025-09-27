package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateAndValidateLicense(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")
	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	handler := NewLicenseHandler(store, "secret")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/licenses":
			handler.CreateLicense(w, r)
		case "/api/v1/licenses/validate":
			handler.ValidateLicense(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := server.Client()

	body, _ := json.Marshal(map[string]string{
		"customerId":  "Acme",
		"fingerprint": "machine",
	})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/licenses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Installer-Token", "secret")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create license request error: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	var createResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	resp.Body.Close()
	key := createResp["licenseKey"]
	if key == "" {
		t.Fatal("expected license key in response")
	}

	valBody, _ := json.Marshal(map[string]string{
		"licenseKey":  key,
		"fingerprint": "machine",
	})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/licenses/validate", bytes.NewReader(valBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Installer-Token", "secret")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("validate request error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestUnauthorizedToken(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")
	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	handler := NewLicenseHandler(store, "secret")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/licenses", bytes.NewReader([]byte("{}")))
	rr := httptest.NewRecorder()

	handler.CreateLicense(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestValidateLicenseExpired(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "licenses.db")
	store, err := NewLicenseStore(dbPath)
	if err != nil {
		t.Fatalf("NewLicenseStore: %v", err)
	}
	defer store.Close()

	handler := NewLicenseHandler(store, "secret")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/licenses":
			handler.CreateLicense(w, r)
		case "/api/v1/licenses/validate":
			handler.ValidateLicense(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := server.Client()

	body, _ := json.Marshal(map[string]string{
		"customerId":  "Acme",
		"fingerprint": "machine",
	})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/licenses", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Installer-Token", "secret")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create license request error: %v", err)
	}
	var createResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	resp.Body.Close()
	key := createResp["licenseKey"]
	if key == "" {
		t.Fatal("expected license key in response")
	}

	if _, err := store.db.ExecContext(context.Background(),
		`UPDATE licenses SET created_at = ? WHERE key = ?`,
		time.Now().UTC().Add(-31*24*time.Hour), key,
	); err != nil {
		t.Fatalf("update license timestamp: %v", err)
	}

	valBody, _ := json.Marshal(map[string]string{
		"licenseKey":  key,
		"fingerprint": "machine",
	})
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/licenses/validate", bytes.NewReader(valBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Installer-Token", "secret")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("validate request error: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	resp.Body.Close()
	if !bytes.Contains(buf.Bytes(), []byte("license expired")) {
		t.Fatalf("expected error message about expiration, got %q", buf.String())
	}
}
