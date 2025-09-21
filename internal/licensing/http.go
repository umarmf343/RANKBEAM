package licensing

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// API exposes HTTP handlers backed by a Service.
type API struct {
	service *Service
}

// NewAPI constructs a new HTTP API wrapper around the provided service.
func NewAPI(service *Service) *API {
	return &API{service: service}
}

// Register attaches the handlers to the supplied mux.
func (api *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/v1/licenses", api.handleIssueLicense)
	mux.HandleFunc("/api/v1/licenses/validate", api.handleValidateLicense)
}

type issueRequest struct {
	Fingerprint string `json:"fingerprint"`
	CustomerID  string `json:"customerId"`
}

type issueResponse struct {
	Status      string     `json:"status"`
	LicenseKey  string     `json:"licenseKey"`
	CustomerID  string     `json:"customerId"`
	Fingerprint string     `json:"fingerprint"`
	IssuedAt    time.Time  `json:"issuedAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

func (api *API) handleIssueLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req issueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	license, err := api.service.IssueLicense(ctx, req.CustomerID, req.Fingerprint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, issueResponse{
		Status:      "issued",
		LicenseKey:  license.Key,
		CustomerID:  license.CustomerID,
		Fingerprint: license.Fingerprint,
		IssuedAt:    license.IssuedAt,
		ExpiresAt:   license.ExpiresAt,
	})
}

type validateRequest struct {
	LicenseKey  string `json:"licenseKey"`
	Fingerprint string `json:"fingerprint"`
}

type validateResponse struct {
	Status      string     `json:"status"`
	LicenseKey  string     `json:"licenseKey"`
	CustomerID  string     `json:"customerId"`
	Fingerprint string     `json:"fingerprint"`
	IssuedAt    time.Time  `json:"issuedAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

func (api *API) handleValidateLicense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req validateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	license, err := api.service.ValidateLicense(ctx, req.LicenseKey, req.Fingerprint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, validateResponse{
		Status:      "valid",
		LicenseKey:  license.Key,
		CustomerID:  license.CustomerID,
		Fingerprint: license.Fingerprint,
		IssuedAt:    license.IssuedAt,
		ExpiresAt:   license.ExpiresAt,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
