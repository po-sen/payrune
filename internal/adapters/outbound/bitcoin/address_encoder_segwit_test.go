package bitcoin

import (
	"testing"

	"payrune/internal/domain/value_objects"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func TestSegwitAddressEncoderScheme(t *testing.T) {
	encoder := NewSegwitAddressEncoder()
	if got := encoder.Scheme(); got != value_objects.BitcoinAddressSchemeSegwit {
		t.Fatalf("unexpected scheme: got %q, want %q", got, value_objects.BitcoinAddressSchemeSegwit)
	}
}

func TestSegwitAddressEncoderEncodeMainnetType(t *testing.T) {
	encoder := NewSegwitAddressEncoder()
	publicKey := newEncoderTestPublicKey(t, &chaincfg.MainNetParams, 11)

	address, err := encoder.Encode(publicKey, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	if !address.IsForNet(&chaincfg.MainNetParams) {
		t.Fatalf("address is not for mainnet: %s", address.EncodeAddress())
	}
	if _, ok := address.(*btcutil.AddressScriptHash); !ok {
		t.Fatalf("unexpected address type: %T", address)
	}
}

func TestSegwitAddressEncoderProvidedVectors(t *testing.T) {
	deriver := newVectorTestDeriver()
	vectors := []providedAddressVector{
		{
			name:     "mainnet segwit",
			network:  value_objects.BitcoinNetworkMainnet,
			scheme:   value_objects.BitcoinAddressSchemeSegwit,
			xpub:     "xpub6BkVzEZTvWGpWSRwjPrfv1UqVQP7s3WUiWE3KU4rkbqtFZrfq3Y9wYPTxbFgAHvToM4kTqETFMzwmxph1CHASZBBjrBKusCRTbjd99i9bdZ",
			expected: "33jyNcCrj627sLyFxLegQWJtHPHHgDF689",
		},
		{
			name:     "testnet4 segwit",
			network:  value_objects.BitcoinNetworkTestnet4,
			scheme:   value_objects.BitcoinAddressSchemeSegwit,
			xpub:     "tpubDCiGGQjmzt8kJU2yX3xQmiL39zKXJKzWDMC3cpoh6Q6VVF1asGd1Je99G72GFLq8t9oMiYPVbmdxwPgKsLspM7kFua2a4NqLQSwAirU2pB2",
			expected: "2N4pZRZ1z84PmKNPuezymytRBFeWfWjHQEN",
		},
	}

	for _, tc := range vectors {
		t.Run(tc.name, func(t *testing.T) {
			assertProvidedVector(t, deriver, tc)
		})
	}
}
