package valueobjects

import "testing"

func TestAddressIssuanceConfigNormalize(t *testing.T) {
	tests := []struct {
		name       string
		input      AddressIssuanceConfig
		wantSource string
		wantPrefix string
	}{
		{
			name: "bitcoin derivation path prefix",
			input: AddressIssuanceConfig{
				AddressSpaceRef:   " xpub-main ",
				IssuanceRefPrefix: "m/84'/0'/0'/",
			},
			wantSource: "xpub-main",
			wantPrefix: "m/84'/0'/0'",
		},
		{
			name: "ethereum create2 namespace",
			input: AddressIssuanceConfig{
				AddressSpaceRef:   " create2.v1:factory=0x1;collector=0x2;init_code_hash=0x3 ",
				IssuanceRefPrefix: "ethereum-mainnet-create2/",
			},
			wantSource: "create2.v1:factory=0x1;collector=0x2;init_code_hash=0x3",
			wantPrefix: "ethereum-mainnet-create2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			normalized := tc.input.Normalize()

			if normalized.AddressSpaceRef != tc.wantSource {
				t.Fatalf("unexpected address source ref: got %q", normalized.AddressSpaceRef)
			}
			if normalized.IssuanceRefPrefix != tc.wantPrefix {
				t.Fatalf("unexpected address reference prefix: got %q", normalized.IssuanceRefPrefix)
			}
		})
	}
}
