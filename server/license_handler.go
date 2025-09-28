package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type LicenseHandler struct {
	store         *LicenseStore
	token         string
	webhookSecret string
	mailer        Mailer
}

func NewLicenseHandler(store *LicenseStore, token, webhookSecret string, mailer Mailer) *LicenseHandler {
	return &LicenseHandler{
		store:         store,
		token:         strings.TrimSpace(token),
		webhookSecret: strings.TrimSpace(webhookSecret),
		mailer:        mailer,
	}
}

func (h *LicenseHandler) HandlePaystackWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}

	if !h.verifySignature(r.Header.Get("x-paystack-signature"), body) {
		http.Error(w, "invalid signature", http.StatusForbidden)
		return
	}

	var payload paystackEvent
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if !payload.IsSuccessfulEvent() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ignored"))
		return
	}

	email := strings.TrimSpace(payload.Data.Customer.Email)
	reference := strings.TrimSpace(payload.Data.Reference)
	if email == "" || reference == "" {
		http.Error(w, "missing customer email or transaction reference", http.StatusBadRequest)
		return
	}

	paidAt := payload.Data.PaidAt.Time
	if payload.Data.PaidAt.IsZero() {
		paidAt = time.Now().UTC()
	}
	expiresAt := paidAt.Add(licenseValidity)

	license, err := h.store.CreateLicense(r.Context(), email, reference, expiresAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.mailer != nil {
		mailCtx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		if err := h.mailer.SendLicenseEmail(mailCtx, email, license); err != nil {
			log.Printf("send license email: %v", err)
		}
		cancel()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"licenseKey": license.Key,
		"expiresAt":  license.ExpiresAt.Format(time.RFC3339),
	})
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
		LicenseKey string `json:"licenseKey"`
		Email      string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	payload.LicenseKey = strings.TrimSpace(payload.LicenseKey)
	payload.Email = strings.TrimSpace(payload.Email)
	if payload.LicenseKey == "" || payload.Email == "" {
		http.Error(w, "licenseKey and email are required", http.StatusBadRequest)
		return
	}

	if _, err := h.store.ValidateLicense(r.Context(), payload.LicenseKey, payload.Email); err != nil {
		switch {
		case errors.Is(err, ErrLicenseNotFound):
			http.Error(w, "license not found", http.StatusUnauthorized)
		case errors.Is(err, ErrEmailMismatch):
			http.Error(w, "email mismatch", http.StatusUnauthorized)
		case errors.Is(err, ErrLicenseExpired):
			http.Error(w, "license expired", http.StatusUnauthorized)
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

func (h *LicenseHandler) verifySignature(signature string, body []byte) bool {
	if h.webhookSecret == "" {
		return true
	}
	mac := hmac.New(sha512.New, []byte(h.webhookSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(strings.TrimSpace(signature)), []byte(expected)) == 1
}

type paystackEvent struct {
	Event string          `json:"event"`
	Data  paystackPayload `json:"data"`
}

func (e paystackEvent) IsSuccessfulEvent() bool {
	if strings.EqualFold(e.Event, "charge.success") || strings.EqualFold(e.Event, "invoice.create") || strings.EqualFold(e.Event, "subscription.create") || strings.EqualFold(e.Event, "subscription.renewed") {
		return true
	}
	return false
}

type paystackPayload struct {
	Reference string           `json:"reference"`
	PaidAt    paystackTime     `json:"paid_at"`
	Customer  paystackCustomer `json:"customer"`
}

type paystackCustomer struct {
	Email string `json:"email"`
}

type paystackTime struct {
	time.Time
}

func (pt *paystackTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		pt.Time = time.Time{}
		return nil
	}
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		pt.Time = time.Time{}
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return err
	}
	pt.Time = parsed
	return nil
}

func (pt paystackTime) IsZero() bool {
	return pt.Time.IsZero()
}
