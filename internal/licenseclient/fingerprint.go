package licenseclient

import (
	"fmt"

	"github.com/denisbrodbeck/machineid"
)

// Fingerprint returns the hashed machine identifier used for licensing.
func Fingerprint() (string, error) {
	id, err := machineid.ID()
	if err != nil {
		return "", fmt.Errorf("obtain machine fingerprint: %w", err)
	}
	return id, nil
}
