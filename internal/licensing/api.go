package licensing

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

// API exposes HTTP handlers for issuing and validating licenses.
type API struct {
	service        *Service
	installerToken string
}

// NewAPI constructs an API wrapper using the provided service. The installer token
// is loaded from the LICENSE_API_TOKEN environment variable; applications can override
// it later via SetInstallerToken.
func NewAPI(service *Service) *API {
	token := strings.TrimSpace(os.Getenv("LICENSE_API_TOKEN"))
	return &API{service: service, installerToken: token}
}

// SetInstallerToken overrides the shared secret expected in the X-Installer-Token header.
func (a *API) SetInstallerToken(token string) {
	a.installerToken = strings.TrimSpace(token)
}

// Register attaches the API handlers to the provided multiplexer.
func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/licenses", a.handleIssue)
	mux.HandleFunc("/api/v1/licenses/validate", a.handleValidate)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func (a *API) handleIssue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.authorize(w, r) {
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

	record, created, err := a.service.IssueLicense(r.Context(), payload.CustomerID, payload.Fingerprint)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"licenseKey": record.Key,
		"issuedAt":   record.IssuedAt.UTC().Format(time.RFC3339),
	}
	if !record.ExpiresAt.IsZero() {
		response["expiresAt"] = record.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if record.CustomerID != "" {
		response["customerId"] = record.CustomerID
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (a *API) handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !a.authorize(w, r) {
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

	record, err := a.service.ValidateLicense(r.Context(), payload.LicenseKey, payload.Fingerprint)
	if err != nil {
		switch {
		case errors.Is(err, ErrLicenseNotFound):
			http.Error(w, "license not found", http.StatusUnauthorized)
		case errors.Is(err, ErrFingerprintMismatch):
			http.Error(w, "fingerprint mismatch", http.StatusUnauthorized)
		case errors.Is(err, ErrLicenseExpired):
			http.Error(w, "license expired", http.StatusUnauthorized)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := map[string]any{
		"status":     "valid",
		"issuedAt":   record.IssuedAt.UTC().Format(time.RFC3339),
		"customerId": record.CustomerID,
	}
	if !record.ExpiresAt.IsZero() {
		response["expiresAt"] = record.ExpiresAt.UTC().Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *API) authorize(w http.ResponseWriter, r *http.Request) bool {
	if strings.TrimSpace(a.installerToken) == "" {
		return true
	}
	provided := r.Header.Get("X-Installer-Token")
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(provided)), []byte(a.installerToken)) == 1 {
		return true
	}
	http.Error(w, "forbidden", http.StatusForbidden)
	return false
}
