package value_objects

import "testing"

func TestParseChainID(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   ChainID
		wantOK bool
	}{
		{name: "trim and lowercase", input: " BitCoin ", want: ChainIDBitcoin, wantOK: true},
		{name: "custom chain value", input: "tron-main", want: ChainID("tron-main"), wantOK: true},
		{name: "with underscore", input: "solana_devnet", want: ChainID("solana_devnet"), wantOK: true},
		{name: "empty", input: " ", want: "", wantOK: false},
		{name: "invalid slash", input: "eth/mainnet", want: "", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseChainID(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("unexpected chain id: got %q, want %q", got, tc.want)
			}
		})
	}
}
