package bitcoin

import (
	"context"
	"errors"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

type fakeIssuedBitcoinAddressDeriver struct {
	address                string
	derivationPath         string
	absoluteDerivationPath string
	err                    error
	lastNetwork            network
	lastScheme             addressScheme
	lastXPub               string
	lastIndex              uint32
}

func (f *fakeIssuedBitcoinAddressDeriver) DeriveAddress(
	network network,
	scheme addressScheme,
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
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          valueobjects.AddressSchemeNativeSegwit,
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
	if output.IssuanceRefKind != valueobjects.IssuanceRefKindHDPathAbsolute {
		t.Fatalf("unexpected issuance ref kind: got %q", output.IssuanceRefKind)
	}
	if output.SweepMaterialJSON != `{"material_type":"bitcoin_hd","material_version":1,"chain":"bitcoin","network":"mainnet","address":"bc1qallocated","hd_derivation_path":"m/84'/0'/0'/0/5","account_xpub":"xpub-main","script_type":"nativeSegwit"}` {
		t.Fatalf("unexpected sweep material json: got %q", output.SweepMaterialJSON)
	}
	if underlying.lastNetwork != networkMainnet {
		t.Fatalf("unexpected network: got %q", underlying.lastNetwork)
	}
	if underlying.lastScheme != addressSchemeNativeSegwit {
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
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: "bitcoin-mainnet-native-segwit",
			Chain:           valueobjects.SupportedChainBitcoin,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          valueobjects.AddressSchemeNativeSegwit,
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
