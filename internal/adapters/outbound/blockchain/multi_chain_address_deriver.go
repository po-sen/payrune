package blockchain

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type chainSpecificAddressDeriver interface {
	Chain() value_objects.SupportedChain
	DeriveAddress(ctx context.Context, input outport.DeriveChainAddressInput) (outport.DeriveChainAddressOutput, error)
}

type MultiChainAddressDeriver struct {
	derivers map[value_objects.SupportedChain]chainSpecificAddressDeriver
}

var _ outport.ChainAddressDeriver = (*MultiChainAddressDeriver)(nil)

func NewMultiChainAddressDeriver(
	derivers ...chainSpecificAddressDeriver,
) (*MultiChainAddressDeriver, error) {
	if len(derivers) == 0 {
		return nil, errors.New("at least one chain address deriver is required")
	}

	normalized := make(map[value_objects.SupportedChain]chainSpecificAddressDeriver, len(derivers))
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

func (d *MultiChainAddressDeriver) SupportsChain(chain value_objects.SupportedChain) bool {
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
		return outport.DeriveChainAddressOutput{}, errors.New("chain is invalid")
	}
	normalizedNetwork, ok := value_objects.ParseNetworkID(string(input.Network))
	if !ok {
		return outport.DeriveChainAddressOutput{}, errors.New("network is invalid")
	}

	deriver, found := d.derivers[normalizedChain]
	if !found {
		return outport.DeriveChainAddressOutput{}, fmt.Errorf(
			"chain address deriver is not configured for chain: %s",
			normalizedChain,
		)
	}

	return deriver.DeriveAddress(ctx, outport.DeriveChainAddressInput{
		Chain:            normalizedChain,
		Network:          normalizedNetwork,
		Scheme:           strings.TrimSpace(input.Scheme),
		AccountPublicKey: strings.TrimSpace(input.AccountPublicKey),
		Index:            input.Index,
	})
}

func normalizeSupportedChain(chain value_objects.SupportedChain) (value_objects.SupportedChain, bool) {
	normalizedChainID, ok := value_objects.ParseChainID(string(chain))
	if !ok {
		return "", false
	}
	return value_objects.SupportedChain(normalizedChainID), true
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
