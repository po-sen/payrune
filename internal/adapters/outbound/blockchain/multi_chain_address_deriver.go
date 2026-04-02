package blockchain

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type chainSpecificAddressDeriver interface {
	Chain() valueobjects.SupportedChain
	DeriveAddress(ctx context.Context, input outport.DeriveChainAddressInput) (outport.DeriveChainAddressOutput, error)
}

type MultiChainAddressDeriver struct {
	derivers map[valueobjects.SupportedChain]chainSpecificAddressDeriver
}

var _ outport.ChainAddressDeriver = (*MultiChainAddressDeriver)(nil)

func NewMultiChainAddressDeriver(
	derivers ...chainSpecificAddressDeriver,
) (*MultiChainAddressDeriver, error) {
	if len(derivers) == 0 {
		return nil, errors.New("at least one chain address deriver is required")
	}

	normalized := make(map[valueobjects.SupportedChain]chainSpecificAddressDeriver, len(derivers))
	for _, deriver := range derivers {
		if isNilChainSpecificAddressDeriver(deriver) {
			return nil, errors.New("chain address deriver is required")
		}

		normalizedChain, ok := normalizeSupportedChain(deriver.Chain())
		if !ok {
			return nil, fmt.Errorf("chain address deriver key is invalid: %s", deriver.Chain())
		}
		if _, exists := normalized[normalizedChain]; exists {
			return nil, fmt.Errorf("chain address deriver is already configured for chain: %s", normalizedChain)
		}

		normalized[normalizedChain] = deriver
	}

	return &MultiChainAddressDeriver{derivers: normalized}, nil
}

func (d *MultiChainAddressDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
	normalizedChain, ok := normalizeSupportedChain(chain)
	if !ok {
		return false
	}

	_, found := d.derivers[normalizedChain]
	return found
}

func (d *MultiChainAddressDeriver) DeriveAddress(
	ctx context.Context,
	input outport.DeriveChainAddressInput,
) (outport.DeriveChainAddressOutput, error) {
	normalizedChain, ok := normalizeSupportedChain(input.Chain)
	if !ok {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}
	normalizedNetwork, ok := valueobjects.ParseNetworkID(string(input.Network))
	if !ok {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}

	deriver, found := d.derivers[normalizedChain]
	if !found {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDeriverNotConfigured
	}

	output, err := deriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:               normalizedChain,
		Network:             normalizedNetwork,
		Scheme:              input.Scheme.Normalize(),
		AddressSpaceRef:     strings.TrimSpace(input.AddressSpaceRef),
		IssuanceRefPrefix:   strings.TrimSpace(input.IssuanceRefPrefix),
		RelativeIssuanceRef: strings.TrimSpace(input.RelativeIssuanceRef),
		SlotIndex:           input.SlotIndex,
	})
	if err != nil {
		switch {
		case errors.Is(err, outport.ErrChainAddressDeriverNotConfigured):
			return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDeriverNotConfigured
		case errors.Is(err, outport.ErrChainAddressDerivationInputInvalid):
			return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
		default:
			return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationFailed
		}
	}
	return output, nil
}

func normalizeSupportedChain(chain valueobjects.SupportedChain) (valueobjects.SupportedChain, bool) {
	normalizedChainID, ok := valueobjects.ParseSupportedChain(string(chain))
	if !ok {
		return "", false
	}
	return normalizedChainID, true
}

func isNilChainSpecificAddressDeriver(deriver chainSpecificAddressDeriver) bool {
	if deriver == nil {
		return true
	}

	value := reflect.ValueOf(deriver)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
