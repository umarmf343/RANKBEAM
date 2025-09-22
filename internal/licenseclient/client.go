package licenseclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/umar/amazon-product-scraper/internal/license"
)

// Envelope captures the details returned from the licensing API.
type Envelope struct {
	LicenseKey  string    `json:"licenseKey"`
	CustomerID  string    `json:"customerId"`
	Fingerprint string    `json:"fingerprint"`
	IssuedAt    time.Time `json:"issuedAt"`
	ExpiresAt   time.Time `json:"expiresAt,omitempty"`
}

// Client wraps the shared license HTTP client used by the installer utilities.
type Client struct {
	inner *license.Client
}

// NewClient constructs a client that communicates with the licensing API hosted at apiBase.
// The LICENSE_API_TOKEN environment variable is automatically forwarded for authentication.
func NewClient(apiBase string) (*Client, error) {
	token := strings.TrimSpace(os.Getenv("LICENSE_API_TOKEN"))
	httpClient := &http.Client{Timeout: 15 * time.Second}
	inner, err := license.NewClient(strings.TrimSpace(apiBase), token, httpClient)
	if err != nil {
		return nil, err
	}
	return &Client{inner: inner}, nil
}

// WithToken overrides the installer token used for subsequent requests.
func (c *Client) WithToken(token string) {
	if c == nil || c.inner == nil {
		return
	}
	c.inner.APIToken = strings.TrimSpace(token)
}

// Issue requests (or reuses) a license for the supplied fingerprint/customer combination.
func (c *Client) Issue(ctx context.Context, fingerprint, customerID string) (*Envelope, error) {
	if c == nil || c.inner == nil {
		return nil, errors.New("licenseclient: client is not initialised")
	}
	fingerprint = strings.TrimSpace(fingerprint)
	customerID = strings.TrimSpace(customerID)
	if fingerprint == "" {
		return nil, errors.New("licenseclient: fingerprint is required")
	}
	if customerID == "" {
		return nil, errors.New("licenseclient: customer identifier is required")
	}

	key, err := c.inner.RequestLicense(ctx, customerID, fingerprint)
	if err != nil {
		return nil, err
	}

	env := &Envelope{
		LicenseKey:  key,
		CustomerID:  customerID,
		Fingerprint: fingerprint,
		IssuedAt:    time.Now().UTC(),
	}
	return env, nil
}

// Fingerprint returns the machine fingerprint used to bind licenses.
func Fingerprint() (string, error) {
	return license.Fingerprint()
}

// Storage persists and retrieves license keys on disk.
type Storage struct {
	path string
}

// NewStorage prepares a storage helper rooted inside the user's configuration directory.
// The appID parameter determines the sub-directory used to isolate licenses for different products.
func NewStorage(appID string) (*Storage, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("licenseclient: resolve config dir: %w", err)
	}
	appDir := filepath.Join(configDir, sanitizeAppID(appID))
	if err := os.MkdirAll(appDir, 0o700); err != nil {
		return nil, fmt.Errorf("licenseclient: create storage dir: %w", err)
	}
	return &Storage{path: filepath.Join(appDir, "license.key")}, nil
}

// Save writes the license key contained in the envelope to disk.
func (s *Storage) Save(env *Envelope) error {
	if s == nil {
		return errors.New("licenseclient: storage is nil")
	}
	if env == nil {
		return errors.New("licenseclient: envelope is nil")
	}
	key := strings.TrimSpace(env.LicenseKey)
	if key == "" {
		return errors.New("licenseclient: empty license key")
	}
	if err := os.WriteFile(s.path, []byte(key+"\n"), 0o600); err != nil {
		return fmt.Errorf("licenseclient: write key: %w", err)
	}
	return nil
}

// Load retrieves the stored license key, trimming whitespace.
func (s *Storage) Load() (string, error) {
	if s == nil {
		return "", errors.New("licenseclient: storage is nil")
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return "", err
	}
	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", errors.New("licenseclient: stored key is empty")
	}
	return key, nil
}

// Path exposes the absolute path to the persisted license key.
func (s *Storage) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func sanitizeAppID(appID string) string {
	appID = strings.TrimSpace(appID)
	if appID == "" {
		return "AmazonProductSuite"
	}
	cleaned := make([]rune, 0, len(appID))
	for _, r := range appID {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-' || r == '_' || r == '.':
			cleaned = append(cleaned, r)
		}
	}
	if len(cleaned) == 0 {
		return "AmazonProductSuite"
	}
	return string(cleaned)
}
