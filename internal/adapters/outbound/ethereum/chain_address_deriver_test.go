package ethereum

import (
	"context"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

func TestChainAddressDeriverChain(t *testing.T) {
	deriver := NewChainAddressDeriver()
	if deriver.Chain() != valueobjects.SupportedChainEthereum {
		t.Fatalf("unexpected chain: got %q", deriver.Chain())
	}
}

func TestChainAddressDeriverDeriveAddressNotImplemented(t *testing.T) {
	deriver := NewChainAddressDeriver()
	_, err := deriver.DeriveAddress(context.Background(), outport.DeriveChainAddressInput{})
	if err == nil || err.Error() != "ethereum deterministic address generation is not implemented" {
		t.Fatalf("unexpected error: got %v", err)
	}
}
