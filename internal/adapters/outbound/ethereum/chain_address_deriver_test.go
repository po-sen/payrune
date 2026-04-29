package ethereum

import (
	"context"
	"testing"

	outport "payrune/internal/application/ports/outbound"
)

func TestBuildCreate2AddressSpaceRef(t *testing.T) {
	sourceRef, err := BuildCreate2AddressSpaceRef(
		"0x1111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222",
		"0x3333333333333333333333333333333333333333333333333333333333333333",
	)
	if err != nil {
		t.Fatalf("BuildCreate2AddressSpaceRef returned error: %v", err)
	}
	if sourceRef != "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333" {
		t.Fatalf("unexpected source ref: got %q", sourceRef)
	}
}

func TestBuildCreate2AddressSpaceRefRejectsInvalidHex(t *testing.T) {
	_, err := BuildCreate2AddressSpaceRef(
		"0x1234",
		"0x2222222222222222222222222222222222222222",
		"0x3333333333333333333333333333333333333333333333333333333333333333",
	)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPredictCreate2AddressMatchesEIP1014Examples(t *testing.T) {
	tests := []struct {
		name           string
		factoryAddress string
		salt           string
		initCode       string
		wantAddress    string
	}{
		{
			name:           "example 0",
			factoryAddress: "0x0000000000000000000000000000000000000000",
			salt:           "0x0000000000000000000000000000000000000000000000000000000000000000",
			initCode:       "0x00",
			wantAddress:    "0x4d1a2e2bb4f88f0250f26ffff098b0b30b26bf38",
		},
		{
			name:           "example 4",
			factoryAddress: "0x00000000000000000000000000000000deadbeef",
			salt:           "0x00000000000000000000000000000000000000000000000000000000cafebabe",
			initCode:       "0xdeadbeef",
			wantAddress:    "0x60f3f640a8508fc6a86d45df051962668e1e8ac7",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, factoryAddress, err := normalizeFixedHex(tc.factoryAddress, 20, "factory address")
			if err != nil {
				t.Fatalf("normalize factory address: %v", err)
			}
			_, saltBytes, err := normalizeFixedHex(tc.salt, 32, "salt")
			if err != nil {
				t.Fatalf("normalize salt: %v", err)
			}
			_, initCode, err := normalizeFixedHex(tc.initCode, len(tc.initCode[2:])/2, "init code")
			if err != nil {
				t.Fatalf("normalize init code: %v", err)
			}

			var salt [32]byte
			copy(salt[:], saltBytes)
			initCodeHash := keccak256Hash(initCode)
			got := predictCreate2Address(factoryAddress, salt, initCodeHash[:])
			if got != tc.wantAddress {
				t.Fatalf("unexpected predicted address: got %q want %q", got, tc.wantAddress)
			}
		})
	}
}

func TestChainAddressDeriverDeriveAddressDeterministically(t *testing.T) {
	deriver := NewChainAddressDeriver()
	input := outport.DeriveChainAddressInput{
		Chain:               outport.SupportedChainEthereum,
		Network:             outport.NetworkIDMainnet,
		Scheme:              "create2",
		AddressSpaceRef:     "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
		RelativeIssuanceRef: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	first, err := deriver.DeriveAddress(context.Background(), input)
	if err != nil {
		t.Fatalf("DeriveAddress returned error: %v", err)
	}
	second, err := deriver.DeriveAddress(context.Background(), input)
	if err != nil {
		t.Fatalf("DeriveAddress returned error on second call: %v", err)
	}

	if first.Address == "" {
		t.Fatal("expected predicted address")
	}
	if first.Address != second.Address {
		t.Fatalf("expected deterministic address, got %q and %q", first.Address, second.Address)
	}
	if first.RelativeIssuanceRef == "" {
		t.Fatal("expected relative address reference")
	}
	if first.IssuanceRefKind != outport.IssuanceRefKindCreate2Salt {
		t.Fatalf("unexpected issuance ref kind: got %q", first.IssuanceRefKind)
	}
	if first.IssuanceRef != first.RelativeIssuanceRef {
		t.Fatalf("unexpected address reference: got %q", first.IssuanceRef)
	}

	other, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{
		Chain:               input.Chain,
		Network:             input.Network,
		Scheme:              input.Scheme,
		AddressSpaceRef:     input.AddressSpaceRef,
		RelativeIssuanceRef: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	})
	if err != nil {
		t.Fatalf("DeriveAddress returned error for different salt: %v", err)
	}
	if other.Address == first.Address {
		t.Fatalf("expected different address for different salt, got %q", other.Address)
	}
}

func TestChainAddressDeriverRejectsInvalidInput(t *testing.T) {
	deriver := NewChainAddressDeriver()

	tests := []outport.DeriveChainAddressInput{
		{
			Chain:               outport.SupportedChainBitcoin,
			Network:             outport.NetworkIDMainnet,
			Scheme:              "create2",
			AddressSpaceRef:     "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
			RelativeIssuanceRef: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			Chain:               outport.SupportedChainEthereum,
			Network:             outport.NetworkIDMainnet,
			Scheme:              "legacy",
			AddressSpaceRef:     "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
			RelativeIssuanceRef: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			Chain:               outport.SupportedChainEthereum,
			Network:             outport.NetworkIDMainnet,
			Scheme:              "create2",
			AddressSpaceRef:     "",
			RelativeIssuanceRef: "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
		{
			Chain:               outport.SupportedChainEthereum,
			Network:             outport.NetworkIDMainnet,
			Scheme:              "create2",
			AddressSpaceRef:     "create2.v1:factory=0x1111111111111111111111111111111111111111;collector=0x2222222222222222222222222222222222222222;init_code_hash=0x3333333333333333333333333333333333333333333333333333333333333333",
			RelativeIssuanceRef: "",
		},
	}

	for _, input := range tests {
		if _, err := deriver.DeriveAddress(context.Background(), input); err == nil {
			t.Fatalf("expected validation error for input: %+v", input)
		}
	}
}
