package bitcoin

import (
	"testing"

	"payrune/internal/domain/value_objects"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

func TestTaprootAddressEncoderScheme(t *testing.T) {
	encoder := NewTaprootAddressEncoder()
	if got := encoder.Scheme(); got != value_objects.BitcoinAddressSchemeTaproot {
		t.Fatalf("unexpected scheme: got %q, want %q", got, value_objects.BitcoinAddressSchemeTaproot)
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
