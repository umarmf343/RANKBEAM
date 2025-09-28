package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func GenerateLicenseKey(email, reference string, expiresAt time.Time) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	reference = strings.TrimSpace(reference)
	if email == "" {
		return "", fmt.Errorf("email is required")
	}
	if reference == "" {
		return "", fmt.Errorf("transaction reference is required")
	}

	emailHash := shortHash(email)
	refHash := shortHash(reference)
	expiry := expiresAt.UTC().Format("2006-01-02")

	return fmt.Sprintf("%s%s-%s", emailHash, refHash, expiry), nil
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	encoded := hex.EncodeToString(sum[:])
	if len(encoded) <= 12 {
		return encoded
	}
	return strings.ToLower(encoded[:12])
}
