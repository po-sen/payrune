package use_cases

import (
	"context"
	"strings"

	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/in"
	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/value_objects"
)

type AddressPolicyConfig struct {
	AddressPolicyID string
	Chain           value_objects.Chain
	Network         value_objects.BitcoinNetwork
	Scheme          value_objects.BitcoinAddressScheme
	MinorUnit       string
	Decimals        uint8
	XPub            string
}

type AddressPolicyCatalog struct {
	ordered []AddressPolicyConfig
	byID    map[string]AddressPolicyConfig
}

func NewAddressPolicyCatalog(configs []AddressPolicyConfig) *AddressPolicyCatalog {
	ordered := make([]AddressPolicyConfig, 0, len(configs))
	byID := make(map[string]AddressPolicyConfig, len(configs))

	for _, cfg := range configs {
		normalized := AddressPolicyConfig{
			AddressPolicyID: strings.TrimSpace(cfg.AddressPolicyID),
			Chain:           cfg.Chain,
			Network:         cfg.Network,
			Scheme:          cfg.Scheme,
			MinorUnit:       strings.TrimSpace(cfg.MinorUnit),
			Decimals:        cfg.Decimals,
			XPub:            strings.TrimSpace(cfg.XPub),
		}

		if normalized.AddressPolicyID == "" {
			continue
		}
		if _, exists := byID[normalized.AddressPolicyID]; exists {
			continue
		}

		ordered = append(ordered, normalized)
		byID[normalized.AddressPolicyID] = normalized
	}

	return &AddressPolicyCatalog{
		ordered: ordered,
		byID:    byID,
	}
}

type listAddressPoliciesUseCase struct {
	catalog *AddressPolicyCatalog
}

func NewListAddressPoliciesUseCase(catalog *AddressPolicyCatalog) inport.ListAddressPoliciesUseCase {
	return &listAddressPoliciesUseCase{catalog: catalog}
}

func (uc *listAddressPoliciesUseCase) Execute(
	_ context.Context,
	chain value_objects.Chain,
) (dto.ListAddressPoliciesResponse, error) {
	if chain != value_objects.ChainBitcoin {
		return dto.ListAddressPoliciesResponse{}, inport.ErrChainNotSupported
	}

	policies := make([]dto.AddressPolicy, 0)
	for _, policy := range uc.catalog.ordered {
		if policy.Chain != chain {
			continue
		}

		policies = append(policies, dto.AddressPolicy{
			AddressPolicyID: policy.AddressPolicyID,
			Chain:           string(policy.Chain),
			Network:         string(policy.Network),
			Scheme:          string(policy.Scheme),
			MinorUnit:       policy.MinorUnit,
			Decimals:        policy.Decimals,
			Enabled:         policy.XPub != "",
		})
	}

	return dto.ListAddressPoliciesResponse{
		Chain:           string(chain),
		AddressPolicies: policies,
	}, nil
}

type generateAddressUseCase struct {
	deriver outport.BitcoinAddressDeriver
	catalog *AddressPolicyCatalog
}

func NewGenerateAddressUseCase(
	deriver outport.BitcoinAddressDeriver,
	catalog *AddressPolicyCatalog,
) inport.GenerateAddressUseCase {
	return &generateAddressUseCase{
		deriver: deriver,
		catalog: catalog,
	}
}

func (uc *generateAddressUseCase) Execute(
	_ context.Context,
	input dto.GenerateAddressInput,
) (dto.GenerateAddressResponse, error) {
	if input.Chain != value_objects.ChainBitcoin {
		return dto.GenerateAddressResponse{}, inport.ErrChainNotSupported
	}

	policy, ok := uc.catalog.byID[strings.TrimSpace(input.AddressPolicyID)]
	if !ok || policy.Chain != input.Chain {
		return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotFound
	}
	if policy.XPub == "" {
		return dto.GenerateAddressResponse{}, inport.ErrAddressPolicyNotEnabled
	}

	address, err := uc.deriver.DeriveAddress(policy.Network, policy.Scheme, policy.XPub, input.Index)
	if err != nil {
		return dto.GenerateAddressResponse{}, err
	}

	return dto.GenerateAddressResponse{
		AddressPolicyID: policy.AddressPolicyID,
		Chain:           string(policy.Chain),
		Network:         string(policy.Network),
		Scheme:          string(policy.Scheme),
		MinorUnit:       policy.MinorUnit,
		Decimals:        policy.Decimals,
		Index:           input.Index,
		Address:         address,
	}, nil
}
