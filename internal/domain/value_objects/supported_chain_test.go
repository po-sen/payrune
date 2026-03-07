package value_objects

import "testing"

func TestParseSupportedChain(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   SupportedChain
		wantOK bool
	}{
		{name: "bitcoin", input: "bitcoin", want: SupportedChainBitcoin, wantOK: true},
		{name: "bitcoin mixed case", input: " BitCoin ", want: SupportedChainBitcoin, wantOK: true},
		{name: "unsupported chain", input: "ethereum", want: "", wantOK: false},
		{name: "invalid identifier", input: "btc/mainnet", want: "", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseSupportedChain(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("unexpected chain: got %q, want %q", got, tc.want)
			}
		})
	}
}
