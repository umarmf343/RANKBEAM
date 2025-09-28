package license

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	ErrMissingBaseURL    = errors.New("license: missing base URL")
	ErrInvalidLicense    = errors.New("license: invalid or expired license")
	ErrUnauthorizedToken = errors.New("license: unauthorized installer token")
	ErrEmptyLicenseKey   = errors.New("license: empty key")
	ErrMissingEmail      = errors.New("license: email address is required")
)

// Client wraps HTTP access to the license server.
type Client struct {
	BaseURL    string
	APIToken   string
	HTTPClient *http.Client
}

// LicenseData represents the persisted license information for a device.
type LicenseData struct {
	Key   string `json:"licenseKey"`
	Email string `json:"email"`
}

// Normalise trims whitespace and lowercases the email address.
func (d LicenseData) Normalise() LicenseData {
	return LicenseData{
		Key:   strings.TrimSpace(d.Key),
		Email: strings.ToLower(strings.TrimSpace(d.Email)),
	}
}

// Validate ensures both the key and email are present.
func (d LicenseData) Validate() error {
	if strings.TrimSpace(d.Key) == "" {
		return ErrEmptyLicenseKey
	}
	if strings.TrimSpace(d.Email) == "" {
		return ErrMissingEmail
	}
	return nil
}

// NewClient constructs a Client and validates the provided base URL.
func NewClient(baseURL, token string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, ErrMissingBaseURL
	}
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return nil, fmt.Errorf("license: base URL must include scheme, got %q", baseURL)
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), APIToken: token, HTTPClient: httpClient}, nil
}

// NewClientFromEnv reads LICENSE_API_URL and LICENSE_API_TOKEN.
func NewClientFromEnv() (*Client, error) {
	base := strings.TrimSpace(os.Getenv("LICENSE_API_URL"))
	token := strings.TrimSpace(os.Getenv("LICENSE_API_TOKEN"))
	return NewClient(base, token, nil)
}

// ValidateLicense ensures the provided license key belongs to the supplied email on the server.
func (c *Client) ValidateLicense(ctx context.Context, licenseKey, email string) error {
	if c == nil {
		return errors.New("license: client is nil")
	}

	payload := LicenseData{Key: licenseKey, Email: email}.Normalise()
	if err := payload.Validate(); err != nil {
		return err
	}

	body, err := json.Marshal(map[string]string{
		"licenseKey": payload.Key,
		"email":      payload.Email,
	})
	if err != nil {
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/api/v1/licenses/validate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result struct {
			Status    string `json:"status"`
			ExpiresAt string `json:"expiresAt,omitempty"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("license: parse response: %w", err)
		}
		if strings.ToLower(strings.TrimSpace(result.Status)) != "valid" {
			return fmt.Errorf("license: unexpected validation status %q", result.Status)
		}
		return nil
	case http.StatusUnauthorized:
		return ErrInvalidLicense
	case http.StatusForbidden:
		return ErrUnauthorizedToken
	default:
		return c.decodeHTTPError(resp)
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	if c == nil {
		return nil, errors.New("license: client is nil")
	}
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("license: parse base URL: %w", err)
	}
	rel, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("license: parse path: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, base.ResolveReference(rel).String(), body)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(c.APIToken) != "" {
		req.Header.Set("X-Installer-Token", c.APIToken)
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) decodeHTTPError(resp *http.Response) error {
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if len(data) == 0 {
		return fmt.Errorf("license: server returned %s", resp.Status)
	}
	return fmt.Errorf("license: server returned %s: %s", resp.Status, strings.TrimSpace(string(data)))
}

// ValidateLocalLicense loads the cached license data and confirms validity with the server.
// It returns the normalised license data on success to allow callers to display or log it.
func ValidateLocalLicense(ctx context.Context, client *Client) (LicenseData, error) {
	if client == nil {
		return LicenseData{}, errors.New("license: client is required")
	}
	data, err := LoadLicense()
	if err != nil {
		return LicenseData{}, err
	}
	if err := client.ValidateLicense(ctx, data.Key, data.Email); err != nil {
		return LicenseData{}, err
	}
	normalised := data.Normalise()
	return normalised, nil
}
