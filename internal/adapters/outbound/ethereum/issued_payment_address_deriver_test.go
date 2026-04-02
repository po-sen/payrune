package ethereum

import (
	"context"
	"encoding/json"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
)

func TestIssuedPaymentAddressDeriverDerivesEthereumCreate2Address(t *testing.T) {
	chainDeriver := NewChainAddressDeriver()
	create2SaltDeriver := NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkIDMainnet: "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	deriver := NewIssuedPaymentAddressDeriver(chainDeriver, create2SaltDeriver)

	output, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: "ethereum-mainnet-create2",
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          "create2",
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
	var sweepMaterial map[string]any
	if err := json.Unmarshal([]byte(output.SweepMaterialJSON), &sweepMaterial); err != nil {
		t.Fatalf("unexpected sweep material json error: %v", err)
	}
	if sweepMaterial["material_type"] != "ethereum_create2" {
		t.Fatalf("unexpected material type: got %v", sweepMaterial["material_type"])
	}
	if sweepMaterial["predicted_address"] != output.Address {
		t.Fatalf("unexpected predicted address: got %v", sweepMaterial["predicted_address"])
	}
	if sweepMaterial["create2_salt"] != output.IssuanceRef {
		t.Fatalf("unexpected create2 salt: got %v", sweepMaterial["create2_salt"])
	}
}

func TestIssuedPaymentAddressDeriverRequiresCreate2SaltDeriver(t *testing.T) {
	deriver := NewIssuedPaymentAddressDeriver(NewChainAddressDeriver(), nil)

	_, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: "ethereum-mainnet-create2",
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          "create2",
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
