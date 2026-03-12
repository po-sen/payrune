package valueobjects

import "testing"

func TestParseNetworkID(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   NetworkID
		wantOK bool
	}{
		{name: "trim and lowercase", input: " MainNet ", want: NetworkID("mainnet"), wantOK: true},
		{name: "custom network", input: "sepolia", want: NetworkID("sepolia"), wantOK: true},
		{name: "with underscore", input: "dev_net", want: NetworkID("dev_net"), wantOK: true},
		{name: "empty", input: " ", want: "", wantOK: false},
		{name: "invalid slash", input: "eth/mainnet", want: "", wantOK: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseNetworkID(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("unexpected network id: got %q, want %q", got, tc.want)
			}
		})
	}
}
