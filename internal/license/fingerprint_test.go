package license

import "testing"

func TestFingerprintFromParts(t *testing.T) {
	parts := []string{"WINDOWS", "AMD64", "HOST", "USER", "AA-BB-CC-DD-EE-FF"}
	fp1 := fingerprintFromParts(parts)
	fp2 := fingerprintFromParts(parts)
	if fp1 != fp2 {
		t.Fatalf("expected deterministic fingerprint, got %q vs %q", fp1, fp2)
	}
	if len(fp1) != 64 {
		t.Fatalf("expected 64 character fingerprint, got %d", len(fp1))
	}
}
