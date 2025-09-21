package license

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"sort"
	"strings"
)

// Fingerprint returns a stable machine fingerprint derived from common hardware
// identifiers. The fingerprint intentionally avoids exposing the raw values by
// hashing the collected components with SHA-256.
func Fingerprint() (string, error) {
	components, err := fingerprintComponents()
	if err != nil {
		return "", err
	}
	if len(components) == 0 {
		return "", fmt.Errorf("license: unable to derive fingerprint components")
	}

	digest := sha256.Sum256([]byte(strings.Join(components, "|")))
	return strings.ToUpper(hex.EncodeToString(digest[:])), nil
}

func fingerprintComponents() ([]string, error) {
	parts := []string{runtime.GOOS, runtime.GOARCH}

	host, err := os.Hostname()
	if err == nil && host != "" {
		parts = append(parts, strings.ToUpper(host))
	}

	currentUser, err := user.Current()
	if err == nil && currentUser.Username != "" {
		parts = append(parts, strings.ToUpper(currentUser.Username))
	}

	macs, err := activeMACAddresses()
	if err != nil {
		return nil, err
	}
	parts = append(parts, macs...)

	return parts, nil
}

func activeMACAddresses() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("license: list interfaces: %w", err)
	}

	macs := make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(iface.HardwareAddr) == 0 {
			continue
		}
		mac := strings.ToUpper(iface.HardwareAddr.String())
		if mac == "" {
			continue
		}
		macs = append(macs, mac)
	}

	sort.Strings(macs)
	return macs, nil
}

func fingerprintFromParts(parts []string) string {
	digest := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return strings.ToUpper(hex.EncodeToString(digest[:]))
}
