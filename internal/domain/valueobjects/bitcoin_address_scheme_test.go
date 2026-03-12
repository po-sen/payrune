package valueobjects

import "testing"

func TestParseBitcoinAddressScheme(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   BitcoinAddressScheme
		wantOK bool
	}{
		{name: "legacy exact", input: "legacy", want: BitcoinAddressSchemeLegacy, wantOK: true},
		{name: "segwit exact", input: "segwit", want: BitcoinAddressSchemeSegwit, wantOK: true},
		{name: "native segwit exact", input: "nativeSegwit", want: BitcoinAddressSchemeNativeSegwit, wantOK: true},
		{name: "taproot exact", input: "taproot", want: BitcoinAddressSchemeTaproot, wantOK: true},
		{name: "native segwit mixed case", input: " NativeSegWit ", want: BitcoinAddressSchemeNativeSegwit, wantOK: true},
		{name: "unsupported", input: "p2pkh", want: "", wantOK: false},
		{name: "empty", input: " ", want: "", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseBitcoinAddressScheme(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("unexpected scheme: got %q, want %q", got, tc.want)
			}
		})
	}
}
