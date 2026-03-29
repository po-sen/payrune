package ethereum

import (
	"context"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

func TestIssuedPaymentAddressDeriverDerivesEthereumCreate2Address(t *testing.T) {
	chainDeriver := NewChainAddressDeriver()
	create2SaltDeriver := NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkID("mainnet"): "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	deriver := NewIssuedPaymentAddressDeriver(chainDeriver, create2SaltDeriver)

	output, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "ethereum-mainnet-create2",
				Chain:           valueobjects.SupportedChainEthereum,
				Network:         valueobjects.NetworkID("mainnet"),
				Scheme:          "create2",
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
				IssuanceRefPrefix: "ethereum-mainnet-create2",
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 145,
			SlotIndex:        11,
		},
	})
	if err != nil {
		t.Fatalf("DeriveIssuedAddress returned error: %v", err)
	}
	if output.Address == "" {
		t.Fatal("expected address")
	}
	if output.IssuanceRef == "" {
		t.Fatal("expected address reference")
	}
}

func TestIssuedPaymentAddressDeriverRequiresCreate2SaltDeriver(t *testing.T) {
	deriver := NewIssuedPaymentAddressDeriver(NewChainAddressDeriver(), nil)

	_, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: "ethereum-mainnet-create2",
				Chain:           valueobjects.SupportedChainEthereum,
				Network:         valueobjects.NetworkID("mainnet"),
				Scheme:          "create2",
			},
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   "configured",
				IssuanceRefPrefix: "ethereum-mainnet-create2",
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 145,
			SlotIndex:        11,
		},
	})
	if err == nil {
		t.Fatal("expected create2 salt deriver error")
	}
}
