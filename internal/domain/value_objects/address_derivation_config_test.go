package value_objects

import "testing"

func TestAddressDerivationConfigNormalize(t *testing.T) {
	config := AddressDerivationConfig{
		AccountPublicKey:         " xpub-main ",
		PublicKeyFingerprintAlgo: " hash160 ",
		PublicKeyFingerprint:     " fingerprint-main ",
		DerivationPathPrefix:     "m/84'/0'/0'/",
	}

	normalized := config.Normalize()

	if normalized.AccountPublicKey != "xpub-main" {
		t.Fatalf("unexpected account public key: got %q", normalized.AccountPublicKey)
	}
	if normalized.PublicKeyFingerprintAlgo != "hash160" {
		t.Fatalf("unexpected fingerprint algorithm: got %q", normalized.PublicKeyFingerprintAlgo)
	}
	if normalized.PublicKeyFingerprint != "fingerprint-main" {
		t.Fatalf("unexpected fingerprint: got %q", normalized.PublicKeyFingerprint)
	}
	if normalized.DerivationPathPrefix != "m/84'/0'/0'" {
		t.Fatalf("unexpected derivation path prefix: got %q", normalized.DerivationPathPrefix)
	}
}

func TestAddressDerivationConfigAbsoluteDerivationPath(t *testing.T) {
	config := AddressDerivationConfig{
		DerivationPathPrefix: "m/84'/0'/0'",
	}

	tests := []struct {
		name     string
		relative string
		want     string
		wantErr  bool
	}{
		{name: "relative", relative: "0/42", want: "m/84'/0'/0'/0/42"},
		{name: "absolute", relative: "m/84'/0'/0'/0/99", want: "m/84'/0'/0'/0/99"},
		{name: "empty", relative: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := config.AbsoluteDerivationPath(tc.relative)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected path: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestAddressDerivationConfigAbsoluteDerivationPathRejectsMissingPrefix(t *testing.T) {
	config := AddressDerivationConfig{}

	if _, err := config.AbsoluteDerivationPath("0/1"); err == nil {
		t.Fatal("expected error when derivation path prefix is missing")
	}
}
