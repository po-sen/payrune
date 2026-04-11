package ethereum

import (
	"context"
	"fmt"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
	ethereumcreate2assets "payrune/internal/infrastructure/ethereumcreate2assets"
)

func TestIssuedPaymentAddressDeriverDerivesEthereumCreate2Address(t *testing.T) {
	chainDeriver := NewChainAddressDeriver()
	create2SaltDeriver := NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkIDMainnet: "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	deriver := NewIssuedPaymentAddressDeriver(chainDeriver, create2SaltDeriver)
	collectorAddress := "0x2222222222222222222222222222222222222222"
	addressSpaceRef := ethereumcreate2assets.BuildAddressSpaceRef("mainnet", collectorAddress)
	if addressSpaceRef == "" {
		t.Fatal("expected native address space ref")
	}

	output, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: "ethereum-mainnet-create2",
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          "create2",
			Enabled:         true,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef:   addressSpaceRef,
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
	if output.IssuanceRefKind != valueobjects.IssuanceRefKindCreate2Salt {
		t.Fatalf("unexpected issuance ref kind: got %q", output.IssuanceRefKind)
	}
	metadata, ok := ethereumcreate2assets.LookupDeploymentMetadata("mainnet")
	if !ok {
		t.Fatal("expected deployment metadata")
	}
	initCodeHex, ok := metadata.Receiver.InitCodeHex(collectorAddress)
	if !ok {
		t.Fatal("expected init code hex")
	}
	initCodeHash, ok := metadata.Receiver.InitCodeHashHex(collectorAddress)
	if !ok {
		t.Fatal("expected init code hash")
	}
	wantSweepMaterial := fmt.Sprintf(
		`{"material_type":"ethereum_create2","material_version":1,"chain":"ethereum","network":"mainnet","address":"%s","predicted_address":"%s","factory_address":"%s","collector_address":"%s","create2_salt":"%s","init_code_hex":"%s","init_code_hash":"%s"}`,
		output.Address,
		output.Address,
		metadata.FactoryAddress,
		collectorAddress,
		output.IssuanceRef,
		initCodeHex,
		initCodeHash,
	)
	if output.SweepMaterial != wantSweepMaterial {
		t.Fatalf("unexpected sweep material:\nwant: %s\n got: %s", wantSweepMaterial, output.SweepMaterial)
	}
}

func TestIssuedPaymentAddressDeriverDerivesEthereumUSDTCreate2Address(t *testing.T) {
	chainDeriver := NewChainAddressDeriver()
	create2SaltDeriver := NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkIDMainnet: "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	deriver := NewIssuedPaymentAddressDeriver(chainDeriver, create2SaltDeriver)
	collectorAddress := "0x2222222222222222222222222222222222222222"
	addressSpaceRef := ethereumcreate2assets.BuildAddressSpaceRef("mainnet", collectorAddress)
	if addressSpaceRef == "" {
		t.Fatal("expected unified address space ref")
	}

	output, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: valueobjects.AddressPolicyIDEthereumMainnetUSDTCreate2,
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          valueobjects.AddressSchemeCreate2,
			AssetReference:  "0xdAC17F958D2ee523a2206206994597C13D831ec7",
			Enabled:         true,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef: addressSpaceRef,
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 245,
			SlotIndex:        12,
		},
	})
	if err != nil {
		t.Fatalf("DeriveIssuedAddress returned error: %v", err)
	}

	tokenReceiverArtifact, ok := ethereumcreate2assets.LookupReceiverArtifact(
		ethereumcreate2assets.ReceiverArtifactName,
	)
	if !ok {
		t.Fatal("expected unified receiver artifact")
	}
	initCodeHex, ok := tokenReceiverArtifact.InitCodeHex(collectorAddress)
	if !ok {
		t.Fatal("expected unified init code hex")
	}
	initCodeHash, ok := tokenReceiverArtifact.InitCodeHashHex(collectorAddress)
	if !ok {
		t.Fatal("expected unified init code hash")
	}
	metadata, ok := ethereumcreate2assets.LookupDeploymentMetadata("mainnet")
	if !ok {
		t.Fatal("expected deployment metadata")
	}

	wantSweepMaterial := fmt.Sprintf(
		`{"material_type":"ethereum_create2","material_version":1,"chain":"ethereum","network":"mainnet","asset_reference":"0xdac17f958d2ee523a2206206994597c13d831ec7","address":"%s","predicted_address":"%s","factory_address":"%s","collector_address":"%s","create2_salt":"%s","init_code_hex":"%s","init_code_hash":"%s"}`,
		output.Address,
		output.Address,
		metadata.FactoryAddress,
		collectorAddress,
		output.IssuanceRef,
		initCodeHex,
		initCodeHash,
	)
	if output.SweepMaterial != wantSweepMaterial {
		t.Fatalf("unexpected usdt sweep material:\nwant: %s\n got: %s", wantSweepMaterial, output.SweepMaterial)
	}
}

func TestIssuedPaymentAddressDeriverKeepsETHAndUSDTAddressesDistinctWithSharedReceiverArtifact(t *testing.T) {
	chainDeriver := NewChainAddressDeriver()
	create2SaltDeriver := NewCreate2SaltDeriver(map[valueobjects.NetworkID]string{
		valueobjects.NetworkIDMainnet: "0x1111111111111111111111111111111111111111111111111111111111111111",
	})
	deriver := NewIssuedPaymentAddressDeriver(chainDeriver, create2SaltDeriver)
	collectorAddress := "0x2222222222222222222222222222222222222222"
	addressSpaceRef := ethereumcreate2assets.BuildAddressSpaceRef("mainnet", collectorAddress)
	if addressSpaceRef == "" {
		t.Fatal("expected unified address space ref")
	}

	nativeOutput, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: valueobjects.AddressPolicyIDEthereumMainnetCreate2,
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          valueobjects.AddressSchemeCreate2,
			Enabled:         true,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef: addressSpaceRef,
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 901,
			SlotIndex:        1,
		},
	})
	if err != nil {
		t.Fatalf("DeriveIssuedAddress native returned error: %v", err)
	}

	usdtOutput, err := deriver.DeriveIssuedAddress(context.Background(), outport.DeriveIssuedPaymentAddressInput{
		Policy: policies.AddressIssuancePolicy{
			AddressPolicyID: valueobjects.AddressPolicyIDEthereumMainnetUSDTCreate2,
			Chain:           valueobjects.SupportedChainEthereum,
			Network:         valueobjects.NetworkIDMainnet,
			Scheme:          valueobjects.AddressSchemeCreate2,
			AssetReference:  "0xdAC17F958D2ee523a2206206994597C13D831ec7",
			Enabled:         true,
			IssuanceConfig: valueobjects.AddressIssuanceConfig{
				AddressSpaceRef: addressSpaceRef,
			},
		}.Normalize(),
		Allocation: entities.PaymentAddressAllocation{
			PaymentAddressID: 901,
			SlotIndex:        1,
		},
	})
	if err != nil {
		t.Fatalf("DeriveIssuedAddress usdt returned error: %v", err)
	}

	if nativeOutput.Address == usdtOutput.Address {
		t.Fatalf("expected ETH and USDT addresses to differ even with unified receiver: got %q", nativeOutput.Address)
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
			Enabled:         true,
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
