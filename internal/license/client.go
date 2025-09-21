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
)

// Client wraps HTTP access to the license server.
type Client struct {
	BaseURL    string
	APIToken   string
	HTTPClient *http.Client
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

// RequestLicense asks the API to issue (or return an existing) license key for
// the supplied fingerprint.
func (c *Client) RequestLicense(ctx context.Context, customerID, fingerprint string) (string, error) {
	payload := map[string]string{
		"customerId":  strings.TrimSpace(customerID),
		"fingerprint": strings.TrimSpace(fingerprint),
	}
	if payload["customerId"] == "" || payload["fingerprint"] == "" {
		return "", fmt.Errorf("license: customer ID and fingerprint are required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/api/v1/licenses", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return "", ErrUnauthorizedToken
	}
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", c.decodeHTTPError(resp)
	}

	var result struct {
		LicenseKey string `json:"licenseKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("license: parse response: %w", err)
	}
	if strings.TrimSpace(result.LicenseKey) == "" {
		return "", fmt.Errorf("license: server returned empty license key")
	}
	return result.LicenseKey, nil
}

// ValidateLicense ensures the provided license key matches the supplied
// fingerprint on the server.
func (c *Client) ValidateLicense(ctx context.Context, licenseKey, fingerprint string) error {
	payload := map[string]string{
		"licenseKey":  strings.TrimSpace(licenseKey),
		"fingerprint": strings.TrimSpace(fingerprint),
	}
	if payload["licenseKey"] == "" || payload["fingerprint"] == "" {
		return fmt.Errorf("license: license key and fingerprint are required")
	}
	body, err := json.Marshal(payload)
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
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("license: parse response: %w", err)
		}
		if strings.ToLower(result.Status) != "valid" {
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
	if c.APIToken != "" {
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

// ValidateLocalLicense loads the cached license key, computes the local
// fingerprint, and confirms validity with the server. It returns the license key
// on success to allow the caller to display or log it.
func ValidateLocalLicense(ctx context.Context, client *Client) (string, error) {
	if client == nil {
		return "", errors.New("license: client is required")
	}
	key, err := LoadLicenseKey()
	if err != nil {
		return "", err
	}
	fingerprint, err := Fingerprint()
	if err != nil {
		return "", err
	}
	if err := client.ValidateLicense(ctx, key, fingerprint); err != nil {
		return "", err
	}
	return key, nil
}
