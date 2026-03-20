package valueobjects

import "testing"

func TestAddressIssuanceConfigNormalize(t *testing.T) {
	config := AddressIssuanceConfig{
		AddressSourceRef:       " xpub-main ",
		AddressReferencePrefix: "m/84'/0'/0'/",
	}

	normalized := config.Normalize()

	if normalized.AddressSourceRef != "xpub-main" {
		t.Fatalf("unexpected address source ref: got %q", normalized.AddressSourceRef)
	}
	if normalized.AddressReferencePrefix != "m/84'/0'/0'" {
		t.Fatalf("unexpected address reference prefix: got %q", normalized.AddressReferencePrefix)
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
