package licenseclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "http://localhost:8080"

// Client communicates with the licensing server.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient builds a client pointing at baseURL. If baseURL is empty, the
// APISUITE_LICENSE_URL environment variable is consulted and falls back to
// defaultBaseURL.
func NewClient(baseURL string) *Client {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = os.Getenv("APISUITE_LICENSE_URL")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Issue requests a new license for the fingerprint/customer pair.
func (c *Client) Issue(ctx context.Context, fingerprint, customerID string) (LicenseEnvelope, error) {
	payload := map[string]string{
		"fingerprint": fingerprint,
		"customerId":  customerID,
	}
	var envelope LicenseEnvelope
	if err := c.post(ctx, "/api/v1/licenses", payload, &envelope); err != nil {
		return LicenseEnvelope{}, err
	}
	return envelope, nil
}

// Validate checks whether the supplied license key matches the fingerprint.
func (c *Client) Validate(ctx context.Context, key, fingerprint string) (LicenseEnvelope, error) {
	payload := map[string]string{
		"licenseKey":  key,
		"fingerprint": fingerprint,
	}
	var envelope LicenseEnvelope
	if err := c.post(ctx, "/api/v1/licenses/validate", payload, &envelope); err != nil {
		return LicenseEnvelope{}, err
	}
	return envelope, nil
}

func (c *Client) post(ctx context.Context, path string, payload any, dest any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Error != "" {
			return errors.New(apiErr.Error)
		}
		return fmt.Errorf("licensing server returned %s", resp.Status)
	}

	if dest != nil {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return err
		}
	}
	return nil
}

// LicenseEnvelope represents the payload returned by the licensing API.
type LicenseEnvelope struct {
	Status      string     `json:"status"`
	LicenseKey  string     `json:"licenseKey"`
	CustomerID  string     `json:"customerId"`
	Fingerprint string     `json:"fingerprint"`
	IssuedAt    time.Time  `json:"issuedAt"`
	ExpiresAt   *time.Time `json:"expiresAt"`
}
