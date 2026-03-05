package value_objects

import "testing"

func TestParseChain(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   Chain
		wantOK bool
	}{
		{name: "bitcoin", input: "bitcoin", want: ChainBitcoin, wantOK: true},
		{name: "bitcoin mixed case", input: " BitCoin ", want: ChainBitcoin, wantOK: true},
		{name: "unsupported chain", input: "ethereum", want: "", wantOK: false},
		{name: "invalid identifier", input: "btc/mainnet", want: "", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseChain(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("unexpected chain: got %q, want %q", got, tc.want)
			}
		})
	}
}
