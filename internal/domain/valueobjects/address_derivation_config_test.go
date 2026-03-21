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
				AddressSourceRef:       " xpub-main ",
				AddressReferencePrefix: "m/84'/0'/0'/",
			},
			wantSource: "xpub-main",
			wantPrefix: "m/84'/0'/0'",
		},
		{
			name: "ethereum create2 namespace",
			input: AddressIssuanceConfig{
				AddressSourceRef:       " create2.v1:factory=0x1;collector=0x2;init_code_hash=0x3 ",
				AddressReferencePrefix: "ethereum-mainnet-create2/",
			},
			wantSource: "create2.v1:factory=0x1;collector=0x2;init_code_hash=0x3",
			wantPrefix: "ethereum-mainnet-create2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			normalized := tc.input.Normalize()

			if normalized.AddressSourceRef != tc.wantSource {
				t.Fatalf("unexpected address source ref: got %q", normalized.AddressSourceRef)
			}
			if normalized.AddressReferencePrefix != tc.wantPrefix {
				t.Fatalf("unexpected address reference prefix: got %q", normalized.AddressReferencePrefix)
			}
		})
	}
}

func TestAddressIssuanceConfigIsEnabled(t *testing.T) {
	if !(AddressIssuanceConfig{AddressSourceRef: " xpub-main "}).IsEnabled() {
		t.Fatal("expected config with address source ref to be enabled")
	}
	if (AddressIssuanceConfig{}).IsEnabled() {
		t.Fatal("expected empty config to be disabled")
	}
}
