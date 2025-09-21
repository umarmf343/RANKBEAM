package main

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
	for _, r := range strings.ToUpper(input) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
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

func HashFingerprint(raw string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(raw)))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func GenerateLicenseKey(customerID, fingerprintHash string) (string, error) {
	sanitized := sanitizeCustomerID(customerID)
	if len(fingerprintHash) < 16 {
		return "", fmt.Errorf("fingerprint hash too short: %d", len(fingerprintHash))
	}
	seg1 := fingerprintHash[0:4]
	seg2 := fingerprintHash[4:8]
	seg3 := fingerprintHash[8:12]
	rand1, err := randomSegment(5)
	if err != nil {
		return "", err
	}
	rand2, err := randomSegment(5)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s-%s", sanitized, seg1, seg2, seg3, rand1, rand2), nil
}

func randomSegment(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid segment length")
	}
	b := make([]rune, length)
	max := len(keyAlphabet)
	for i := range b {
		n, err := rand.Int(rand.Reader, bigInt(max))
		if err != nil {
			return "", err
		}
		b[i] = keyAlphabet[n.Int64()]
	}
	return string(b), nil
}

func bigInt(n int) *big.Int {
	return big.NewInt(int64(n))
}
