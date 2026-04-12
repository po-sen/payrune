package blockchain

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

type chainSpecificIssuedPaymentAddressDeriver interface {
	Chain() valueobjects.SupportedChain
	DeriveIssuedAddress(
		ctx context.Context,
		input outport.DeriveIssuedPaymentAddressInput,
	) (outport.DeriveIssuedPaymentAddressOutput, error)
}

type MultiChainIssuedPaymentAddressDeriver struct {
	derivers map[valueobjects.SupportedChain]chainSpecificIssuedPaymentAddressDeriver
}

var _ outport.IssuedPaymentAddressDeriver = (*MultiChainIssuedPaymentAddressDeriver)(nil)

func NewMultiChainIssuedPaymentAddressDeriver(
	derivers ...chainSpecificIssuedPaymentAddressDeriver,
) (*MultiChainIssuedPaymentAddressDeriver, error) {
	if len(derivers) == 0 {
		return nil, errors.New("at least one issued payment address deriver is required")
	}

	normalized := make(map[valueobjects.SupportedChain]chainSpecificIssuedPaymentAddressDeriver, len(derivers))
	for _, deriver := range derivers {
		if isNilChainSpecificIssuedPaymentAddressDeriver(deriver) {
			return nil, errors.New("issued payment address deriver is required")
		}

		normalizedChain, ok := normalizeSupportedChain(deriver.Chain())
		if !ok {
			return nil, fmt.Errorf("issued payment address deriver chain key is invalid: %s", deriver.Chain())
		}
		if _, exists := normalized[normalizedChain]; exists {
			return nil, fmt.Errorf("issued payment address deriver is already configured for chain: %s", normalizedChain)
		}

		normalized[normalizedChain] = deriver
	}

	return &MultiChainIssuedPaymentAddressDeriver{derivers: normalized}, nil
}

func (d *MultiChainIssuedPaymentAddressDeriver) SupportsChain(chain valueobjects.SupportedChain) bool {
	normalizedChain, ok := normalizeSupportedChain(chain)
	if !ok {
		return false
	}

	_, found := d.derivers[normalizedChain]
	return found
}

func (d *MultiChainIssuedPaymentAddressDeriver) DeriveIssuedAddress(
	ctx context.Context,
	input outport.DeriveIssuedPaymentAddressInput,
) (outport.DeriveIssuedPaymentAddressOutput, error) {
	policy := input.Policy.Normalize()
	normalizedChain, ok := normalizeSupportedChain(policy.Chain)
	if !ok {
		return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDerivationInputInvalid
	}

	deriver, found := d.derivers[normalizedChain]
	if !found {
		return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDeriverNotConfigured
	}

	output, err := deriver.DeriveIssuedAddress(ctx, outport.DeriveIssuedPaymentAddressInput{
		Policy:     policy,
		Allocation: input.Allocation,
	})
	if err != nil {
		switch {
		case errors.Is(err, outport.ErrIssuedPaymentAddressDeriverNotConfigured):
			return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDeriverNotConfigured
		case errors.Is(err, outport.ErrIssuedPaymentAddressDerivationInputInvalid):
			return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDerivationInputInvalid
		default:
			return outport.DeriveIssuedPaymentAddressOutput{}, outport.ErrIssuedPaymentAddressDerivationFailed
		}
	}
	return output, nil
}

func isNilChainSpecificIssuedPaymentAddressDeriver(deriver chainSpecificIssuedPaymentAddressDeriver) bool {
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

func normalizeSupportedChain(chain valueobjects.SupportedChain) (valueobjects.SupportedChain, bool) {
	return valueobjects.ParseSupportedChain(string(chain))
}
