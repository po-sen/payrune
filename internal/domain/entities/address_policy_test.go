package entities

import "testing"

func TestAddressPolicyAbsoluteDerivationPath(t *testing.T) {
	policy := AddressPolicy{
		AddressPolicyID:      "bitcoin-mainnet-native-segwit",
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
			got, err := policy.AbsoluteDerivationPath(tc.relative)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected path: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAddressPolicyAbsoluteDerivationPathRejectMissingPrefix(t *testing.T) {
	policy := AddressPolicy{
		AddressPolicyID:      "bitcoin-mainnet-native-segwit",
		DerivationPathPrefix: "",
	}

	if _, err := policy.AbsoluteDerivationPath("0/1"); err == nil {
		t.Fatalf("expected error when derivation path prefix is missing")
	}
}
