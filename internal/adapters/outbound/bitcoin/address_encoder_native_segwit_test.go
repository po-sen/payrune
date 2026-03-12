package bitcoin

import (
	"testing"

	"payrune/internal/domain/valueobjects"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func TestNativeSegwitAddressEncoderScheme(t *testing.T) {
	encoder := NewNativeSegwitAddressEncoder()
	if got := encoder.Scheme(); got != valueobjects.BitcoinAddressSchemeNativeSegwit {
		t.Fatalf("unexpected scheme: got %q, want %q", got, valueobjects.BitcoinAddressSchemeNativeSegwit)
	}
}

func TestNativeSegwitAddressEncoderEncodeMainnetType(t *testing.T) {
	encoder := NewNativeSegwitAddressEncoder()
	publicKey := newEncoderTestPublicKey(t, &chaincfg.MainNetParams, 11)

	address, err := encoder.Encode(publicKey, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	if !address.IsForNet(&chaincfg.MainNetParams) {
		t.Fatalf("address is not for mainnet: %s", address.EncodeAddress())
	}
	if _, ok := address.(*btcutil.AddressWitnessPubKeyHash); !ok {
		t.Fatalf("unexpected address type: %T", address)
	}
}

func TestNativeSegwitAddressEncoderProvidedVectors(t *testing.T) {
	deriver := newVectorTestDeriver()
	vectors := []providedAddressVector{
		{
			name:     "mainnet native segwit",
			network:  valueobjects.BitcoinNetworkMainnet,
			scheme:   valueobjects.BitcoinAddressSchemeNativeSegwit,
			xpub:     "xpub6DFsnqJG199XeaNU1L4oamyEJkDi8ZkKY6KopjptkMGhFLUSu8SGVYY6TJm9Yz8i6eRVkUCwKUTYHBo7UFqdBaSkb1takgPdcAQK8e6ZjQV",
			expected: "bc1qh07g837l8k2dnh5rxaeq36vhz7funkrr9zsx5t",
		},
		{
			name:     "testnet4 native segwit",
			network:  valueobjects.BitcoinNetworkTestnet4,
			scheme:   valueobjects.BitcoinAddressSchemeNativeSegwit,
			xpub:     "tpubDCiB1iLoNxaaj4MTSk2DoTuwUpEfgm4E3vAcTnvG64rR1smhcEsoTeqNCB4af1XHGspgNfWBA3ccpXiwX5JtxwZMTFct6DQWzrKundqdwEq",
			expected: "tb1qc9k5y2v7r57gg49jcm8ct6m0utru69dav59796",
		},
	}

	for _, tc := range vectors {
		t.Run(tc.name, func(t *testing.T) {
			assertProvidedVector(t, deriver, tc)
		})
	}
}
