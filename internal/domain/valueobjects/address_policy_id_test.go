package valueobjects

import "testing"

func TestNewAddressPolicyID(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   AddressPolicyID
		wantOK bool
	}{
		{
			name:   "canonical bitcoin policy",
			input:  "bitcoin-mainnet-native-segwit",
			want:   AddressPolicyIDBitcoinMainnetNativeSegwit,
			wantOK: true,
		},
		{
			name:   "trim and lowercase",
			input:  " Ethereum-Sepolia-Create2 ",
			want:   AddressPolicyIDEthereumSepoliaCreate2,
			wantOK: true,
		},
		{
			name:   "reject empty",
			input:  "   ",
			wantOK: false,
		},
		{
			name:   "reject invalid rune",
			input:  "policy/1",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewAddressPolicyID(tc.input)
			ok := err == nil
			if ok != tc.wantOK {
				t.Fatalf("unexpected ok: got %v, want %v (err=%v)", ok, tc.wantOK, err)
			}
			if got != tc.want {
				t.Fatalf("unexpected id: got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAddressPolicyIDNormalize(t *testing.T) {
	id := AddressPolicyID(" Bitcoin-Mainnet-Legacy ")
	if got := id.Normalize(); got != AddressPolicyIDBitcoinMainnetLegacy {
		t.Fatalf("unexpected normalized id: got %q", got)
	}

	if !AddressPolicyID(" ").IsZero() {
		t.Fatalf("expected empty id to be zero")
	}
}

func TestEthereumCreate2AddressPolicyID(t *testing.T) {
	tests := []struct {
		name    string
		network NetworkID
		want    AddressPolicyID
	}{
		{
			name:    "built in mainnet",
			network: NetworkIDMainnet,
			want:    AddressPolicyIDEthereumMainnetCreate2,
		},
		{
			name:    "built in sepolia",
			network: NetworkIDSepolia,
			want:    AddressPolicyIDEthereumSepoliaCreate2,
		},
		{
			name:    "open ended valid network",
			network: NetworkID("holesky"),
			want:    AddressPolicyID("ethereum-holesky-create2"),
		},
		{
			name:    "invalid network",
			network: NetworkID("eth/mainnet"),
			want:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := EthereumCreate2AddressPolicyID(tc.network); got != tc.want {
				t.Fatalf("unexpected policy id: got %q, want %q", got, tc.want)
			}
		})
	}
}
