package bitcoin

import (
	"testing"

	"payrune/internal/domain/value_objects"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func TestLegacyAddressEncoderScheme(t *testing.T) {
	encoder := NewLegacyAddressEncoder()
	if got := encoder.Scheme(); got != value_objects.BitcoinAddressSchemeLegacy {
		t.Fatalf("unexpected scheme: got %q, want %q", got, value_objects.BitcoinAddressSchemeLegacy)
	}
}

func TestLegacyAddressEncoderEncodeMainnetType(t *testing.T) {
	encoder := NewLegacyAddressEncoder()
	publicKey := newEncoderTestPublicKey(t, &chaincfg.MainNetParams, 11)

	address, err := encoder.Encode(publicKey, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatalf("Encode returned error: %v", err)
	}
	if !address.IsForNet(&chaincfg.MainNetParams) {
		t.Fatalf("address is not for mainnet: %s", address.EncodeAddress())
	}
	if _, ok := address.(*btcutil.AddressPubKeyHash); !ok {
		t.Fatalf("unexpected address type: %T", address)
	}
}

func TestLegacyAddressEncoderProvidedVectors(t *testing.T) {
	deriver := newVectorTestDeriver()
	vectors := []providedAddressVector{
		{
			name:     "mainnet legacy",
			network:  value_objects.BitcoinNetworkMainnet,
			scheme:   value_objects.BitcoinAddressSchemeLegacy,
			xpub:     "xpub6DUmTFpDUjs36einPToqDQXgUNXWZgUP7TFxsP1ToiBqmbNyyNusGnEdjKfaBr7TL66E9AoEWY6ap5Ra7a5cC6scE4gG4u31fJL1HFmrQ2a",
			expected: "1KpBQPxVLPqPfPvF9ozignj53hPJEMQEmw",
		},
		{
			name:     "testnet4 legacy",
			network:  value_objects.BitcoinNetworkTestnet4,
			scheme:   value_objects.BitcoinAddressSchemeLegacy,
			xpub:     "tpubDDoLYVq7AUqYP63QvYZxnxk1pCJnWDWdzu9w3BYTP9dJAX47xknZiEKUheaAahn6zBNT5ndCzY2x6MQ8iVj7QpFwuhm5bDF6Ggt3q1Rn2Qs",
			expected: "mtFJ951QbSd5FtBsD8JSfSnizqHBJY3SAB",
		},
	}

	for _, tc := range vectors {
		t.Run(tc.name, func(t *testing.T) {
			assertProvidedVector(t, deriver, tc)
		})
	}
}
