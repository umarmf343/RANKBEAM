package licensing

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

var keyAlphabet = []rune("ABCDEFGHJKLMNPQRSTUVWXYZ23456789")

func sanitizeCustomerID(input string) string {
	cleaned := make([]rune, 0, len(input))
	for _, r := range strings.ToUpper(strings.TrimSpace(input)) {
		switch {
		case r >= 'A' && r <= 'Z':
			cleaned = append(cleaned, r)
		case r >= '0' && r <= '9':
			cleaned = append(cleaned, r)
		}
	}
	if len(cleaned) == 0 {
		return "CUSTOMER"
	}
	if len(cleaned) > 12 {
		cleaned = cleaned[:12]
	}
	return string(cleaned)
}

// HashFingerprint normalises and hashes the fingerprint string.
func HashFingerprint(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

// GenerateLicenseKey composes a deterministic but hard-to-guess license key using
// the fingerprint hash and customer identifier.
func GenerateLicenseKey(customerID, fingerprintHash string) (string, error) {
	sanitized := sanitizeCustomerID(customerID)
	if len(fingerprintHash) < 12 {
		return "", fmt.Errorf("licensing: fingerprint hash too short (%d)", len(fingerprintHash))
	}
	segment := func(start, end int) string { return fingerprintHash[start:end] }
	rand1, err := randomSegment(5)
	if err != nil {
		return "", err
	}
	rand2, err := randomSegment(5)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s",
		sanitized,
		segment(0, 4),
		segment(4, 8),
		segment(8, 12),
		rand1,
		rand2,
	), nil
}

func randomSegment(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("licensing: invalid segment length")
	}
	buf := make([]rune, length)
	max := big.NewInt(int64(len(keyAlphabet)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("licensing: random segment: %w", err)
		}
		buf[i] = keyAlphabet[int(n.Int64())]
	}
	return string(buf), nil
}
