package bitcoin

import (
	"testing"

	"payrune/internal/domain/valueobjects"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func TestTaprootAddressEncoderScheme(t *testing.T) {
	encoder := NewTaprootAddressEncoder()
	if got := encoder.Scheme(); got != valueobjects.BitcoinAddressSchemeTaproot {
		t.Fatalf("unexpected scheme: got %q, want %q", got, valueobjects.BitcoinAddressSchemeTaproot)
	}
}

func TestTaprootAddressEncoderEncodeMainnetType(t *testing.T) {
	encoder := NewTaprootAddressEncoder()
	publicKey := newEncoderTestPublicKey(t, &chaincfg.MainNetParams, 11)

	address, err := encoder.Encode(publicKey, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	if !address.IsForNet(&chaincfg.MainNetParams) {
		t.Fatalf("address is not for mainnet: %s", address.EncodeAddress())
	}
	if _, ok := address.(*btcutil.AddressTaproot); !ok {
		t.Fatalf("unexpected address type: %T", address)
	}
}

func TestTaprootAddressEncoderProvidedVectors(t *testing.T) {
	deriver := newVectorTestDeriver()
	vectors := []providedAddressVector{
		{
			name:     "mainnet taproot",
			network:  valueobjects.BitcoinNetworkMainnet,
			scheme:   valueobjects.BitcoinAddressSchemeTaproot,
			xpub:     "xpub6BmoyGVa8shrEFn34McKpK8fkEXijKSuhXQbt4UsJzbZWRrLBkxJLptnuvivSbZA2zWBxvHFgaLs1iB9PMH9Frnse8jpNzZB8Q4k6hFw9c6",
			expected: "bc1pu3k4ewj2a6gsllcfjpge5hdg52gsaljdpzmufjgx00xkgp2alfnq7v330g",
		},
		{
			name:     "testnet4 taproot",
			network:  valueobjects.BitcoinNetworkTestnet4,
			scheme:   valueobjects.BitcoinAddressSchemeTaproot,
			xpub:     "tpubDCEp3dYAyqnXrXPDSw4bhmj6cB6KmM2SkpDXzJWmQ595tPFobRSxhajfz7Yq5ZJvZaQ2qzQDWgaFiihSEQQJn12qtweLnveaA7FYLcpcF97",
			expected: "tb1pzwwf9c7vavp45647p7eg5fe64xm4u4qcwzvsaudw8a8n0hmvpm5ssqd7um",
		},
	}

	for _, tc := range vectors {
		t.Run(tc.name, func(t *testing.T) {
			assertProvidedVector(t, deriver, tc)
		})
	}
}
