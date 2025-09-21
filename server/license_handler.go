package main

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type LicenseHandler struct {
	store *LicenseStore
	token string
}

func NewLicenseHandler(store *LicenseStore, token string) *LicenseHandler {
	return &LicenseHandler{store: store, token: strings.TrimSpace(token)}
}

func (h *LicenseHandler) CreateLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.authorize(w, r) {
		return
	}

	var payload struct {
		CustomerID  string `json:"customerId"`
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	payload.CustomerID = strings.TrimSpace(payload.CustomerID)
	payload.Fingerprint = strings.TrimSpace(payload.Fingerprint)
	if payload.CustomerID == "" || payload.Fingerprint == "" {
		http.Error(w, "customerId and fingerprint are required", http.StatusBadRequest)
		return
	}

	hash := HashFingerprint(payload.Fingerprint)
	sanitized := sanitizeCustomerID(payload.CustomerID)
	license, created, err := h.store.CreateLicense(r.Context(), sanitized, hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"licenseKey": license.Key})
}

func (h *LicenseHandler) ValidateLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.authorize(w, r) {
		return
	}

	var payload struct {
		LicenseKey  string `json:"licenseKey"`
		Fingerprint string `json:"fingerprint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	payload.LicenseKey = strings.TrimSpace(payload.LicenseKey)
	payload.Fingerprint = strings.TrimSpace(payload.Fingerprint)
	if payload.LicenseKey == "" || payload.Fingerprint == "" {
		http.Error(w, "licenseKey and fingerprint are required", http.StatusBadRequest)
		return
	}

	hash := HashFingerprint(payload.Fingerprint)
	if _, err := h.store.ValidateLicense(r.Context(), payload.LicenseKey, hash); err != nil {
		switch {
		case errors.Is(err, ErrLicenseNotFound):
			http.Error(w, "license not found", http.StatusUnauthorized)
		case errors.Is(err, ErrFingerprintMismatch):
			http.Error(w, "fingerprint mismatch", http.StatusUnauthorized)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "valid"})
}

func (h *LicenseHandler) authorize(w http.ResponseWriter, r *http.Request) bool {
	if h.token == "" {
		return true
	}
	provided := r.Header.Get("X-Installer-Token")
	if subtle.ConstantTimeCompare([]byte(provided), []byte(h.token)) == 1 {
		return true
	}
	http.Error(w, "forbidden", http.StatusForbidden)
	return false
}
