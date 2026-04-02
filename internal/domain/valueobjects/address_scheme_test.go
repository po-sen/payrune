package valueobjects

import "testing"

func TestParseAddressScheme(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   AddressScheme
		wantOK bool
	}{
		{name: "legacy", input: "legacy", want: AddressSchemeLegacy, wantOK: true},
		{name: "segwit", input: "segwit", want: AddressSchemeSegwit, wantOK: true},
		{name: "native segwit mixed case", input: " NativeSegWit ", want: AddressSchemeNativeSegwit, wantOK: true},
		{name: "taproot", input: "taproot", want: AddressSchemeTaproot, wantOK: true},
		{name: "create2", input: " create2 ", want: AddressSchemeCreate2, wantOK: true},
		{name: "reject unknown", input: "weird", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseAddressScheme(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("unexpected scheme: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAddressSchemeNormalize(t *testing.T) {
	if got := AddressScheme(" NativeSegWit ").Normalize(); got != AddressSchemeNativeSegwit {
		t.Fatalf("unexpected canonical scheme: got %q", got)
	}
	if got := AddressScheme(" custom ").Normalize(); got != AddressScheme("custom") {
		t.Fatalf("unexpected trimmed custom scheme: got %q", got)
	}
	if AddressScheme(" ").Normalize() != "" {
		t.Fatalf("expected blank scheme to normalize to zero")
	}
	if !AddressSchemeCreate2.IsCreate2() {
		t.Fatalf("expected create2 helper to detect create2")
	}
}
