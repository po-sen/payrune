package bitcoin

import (
	"context"
	"errors"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

type fakeIssuedBitcoinAddressDeriver struct {
	address                string
	derivationPath         string
	absoluteDerivationPath string
	err                    error
	lastNetwork            valueobjects.BitcoinNetwork
	lastScheme             valueobjects.BitcoinAddressScheme
	lastXPub               string
	lastIndex              uint32
}

func (f *fakeIssuedBitcoinAddressDeriver) DeriveAddress(
	network valueobjects.BitcoinNetwork,
	scheme valueobjects.BitcoinAddressScheme,
	xpub string,
	index uint32,
) (string, error) {
	f.lastNetwork = network
	f.lastScheme = scheme
	f.lastXPub = xpub
	f.lastIndex = index
	if f.err != nil {
		return "", f.err
	}
	return f.address, nil
}

func (f *fakeIssuedBitcoinAddressDeriver) DerivationPath(_ string, _ uint32) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.derivationPath, nil
}

func (f *fakeIssuedBitcoinAddressDeriver) AbsoluteDerivationPath(_ string, _ string, _ uint32) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.absoluteDerivationPath, nil
}

func TestIssuedPaymentAddressDeriverDerivesBitcoinAddress(t *testing.T) {
	underlying := &fakeIssuedBitcoinAddressDeriver{
		address:                "bc1qallocated",
		derivationPath:         "0/5",
		absoluteDerivationPath: "m/84'/0'/0'/0/5",
	}
	deriver := NewIssuedPaymentAddressDeriver(NewChainAddressDeriver(underlying))

	output, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "bitcoin-mainnet-native-segwit",
				Chain:           valueobjects.SupportedChainBitcoin,
				Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
				Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "xpub-main",
				IssuanceRefPrefix: "m/84'/0'/0'",
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 55,
			SlotIndex:        5,
		},
	})
	if err != nil {
		t.Fatalf("DeriveIssuedAddress returned error: %v", err)
	}
	if output.Address != "bc1qallocated" {
		t.Fatalf("unexpected address: got %q", output.Address)
	}
	if output.IssuanceRef != "m/84'/0'/0'/0/5" {
		t.Fatalf("unexpected address reference: got %q", output.IssuanceRef)
	}
	if underlying.lastNetwork != valueobjects.BitcoinNetworkMainnet {
		t.Fatalf("unexpected network: got %q", underlying.lastNetwork)
	}
	if underlying.lastScheme != valueobjects.BitcoinAddressSchemeNativeSegwit {
		t.Fatalf("unexpected scheme: got %q", underlying.lastScheme)
	}
	if underlying.lastXPub != "xpub-main" {
		t.Fatalf("unexpected xpub: got %q", underlying.lastXPub)
	}
	if underlying.lastIndex != 5 {
		t.Fatalf("unexpected index: got %d", underlying.lastIndex)
	}
}

func TestIssuedPaymentAddressDeriverPropagatesChainDeriverError(t *testing.T) {
	deriver := NewIssuedPaymentAddressDeriver(NewChainAddressDeriver(&fakeIssuedBitcoinAddressDeriver{
		err: errors.New("derive failed"),
	}))

	_, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "bitcoin-mainnet-native-segwit",
				Chain:           valueobjects.SupportedChainBitcoin,
				Network:         valueobjects.NetworkID(valueobjects.BitcoinNetworkMainnet),
				Scheme:          string(valueobjects.BitcoinAddressSchemeNativeSegwit),
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "xpub-main",
				IssuanceRefPrefix: "m/84'/0'/0'",
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 44,
			SlotIndex:        11,
		},
	})
	if !errors.Is(err, outport.ErrIssuedPaymentAddressDerivationFailed) {
		t.Fatalf("expected %v, got %v", outport.ErrIssuedPaymentAddressDerivationFailed, err)
	}
}
