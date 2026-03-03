package value_objects

import "testing"

func TestParseBitcoinNetwork(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   BitcoinNetwork
		wantOK bool
	}{
		{name: "mainnet exact", input: "mainnet", want: BitcoinNetworkMainnet, wantOK: true},
		{name: "testnet4 exact", input: "testnet4", want: BitcoinNetworkTestnet4, wantOK: true},
		{name: "mainnet mixed case", input: " MainNet ", want: BitcoinNetworkMainnet, wantOK: true},
		{name: "testnet4 upper case", input: " TESTNET4 ", want: BitcoinNetworkTestnet4, wantOK: true},
		{name: "unsupported", input: "regtest", want: "", wantOK: false},
		{name: "empty", input: " ", want: "", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseBitcoinNetwork(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("unexpected network: got %q, want %q", got, tc.want)
			}
		})
	}
}
