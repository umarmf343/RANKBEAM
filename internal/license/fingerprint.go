package license

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/denisbrodbeck/machineid"
)

// HardwareFingerprint returns a stable, anonymised identifier for the current machine.
func HardwareFingerprint() (string, error) {
	id, err := machineid.ProtectedID("RankBeam")
	if err != nil {
		id, err = machineid.ID()
		if err != nil {
			return "", fmt.Errorf("license: resolve hardware fingerprint: %w", err)
		}
	}
	sum := sha256.Sum256([]byte(id))
	fingerprint := strings.ToUpper(hex.EncodeToString(sum[:]))
	return fingerprint[:32], nil
}
