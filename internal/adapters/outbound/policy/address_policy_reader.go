package policy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/valueobjects"
)

const accountPublicKeyFingerprintAlgorithmSHA256Trunc64HexV1 = "sha256-trunc64-hex-v1"

type AddressPolicyConfig struct {
	AddressPolicyID          string
	Chain                    valueobjects.SupportedChain
	Network                  valueobjects.NetworkID
	Scheme                   string
	MinorUnit                string
	Decimals                 uint8
	AccountPublicKey         string
	PublicKeyFingerprintAlgo string
	PublicKeyFingerprint     string
	DerivationPathPrefix     string
}

type addressPolicyReader struct {
	ordered      []entities.AddressPolicy
	issuanceByID map[string]entities.AddressIssuancePolicy
}

var _ outport.AddressPolicyReader = (*addressPolicyReader)(nil)

func NewAddressPolicyReader(configs []AddressPolicyConfig) outport.AddressPolicyReader {
	ordered := make([]entities.AddressPolicy, 0, len(configs))
	issuanceByID := make(map[string]entities.AddressIssuancePolicy, len(configs))

	for _, cfg := range configs {
		issuancePolicy := entities.AddressIssuancePolicy{
			AddressPolicy: entities.AddressPolicy{
				AddressPolicyID: strings.TrimSpace(cfg.AddressPolicyID),
				Chain:           cfg.Chain,
				Network:         cfg.Network,
				Scheme:          cfg.Scheme,
				MinorUnit:       strings.TrimSpace(cfg.MinorUnit),
				Decimals:        cfg.Decimals,
			},
			DerivationConfig: valueobjects.AddressDerivationConfig{
				AccountPublicKey:         strings.TrimSpace(cfg.AccountPublicKey),
				PublicKeyFingerprintAlgo: strings.TrimSpace(cfg.PublicKeyFingerprintAlgo),
				PublicKeyFingerprint:     strings.TrimSpace(cfg.PublicKeyFingerprint),
				DerivationPathPrefix:     strings.TrimSpace(cfg.DerivationPathPrefix),
			},
		}.Normalize()

		if issuancePolicy.DerivationConfig.IsEnabled() && issuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo == "" {
			issuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo = accountPublicKeyFingerprintAlgorithmSHA256Trunc64HexV1
		}
		if issuancePolicy.DerivationConfig.IsEnabled() && issuancePolicy.DerivationConfig.PublicKeyFingerprint == "" {
			issuancePolicy.DerivationConfig.PublicKeyFingerprintAlgo = accountPublicKeyFingerprintAlgorithmSHA256Trunc64HexV1
			issuancePolicy.DerivationConfig.PublicKeyFingerprint = fingerprintAccountPublicKeySHA256Trunc64HexV1(
				issuancePolicy.DerivationConfig.AccountPublicKey,
			)
		}
		issuancePolicy = issuancePolicy.Normalize()

		if issuancePolicy.AddressPolicy.AddressPolicyID == "" {
			continue
		}
		if _, exists := issuanceByID[issuancePolicy.AddressPolicy.AddressPolicyID]; exists {
			continue
		}

		ordered = append(ordered, issuancePolicy.AddressPolicy)
		issuanceByID[issuancePolicy.AddressPolicy.AddressPolicyID] = issuancePolicy
	}

	return &addressPolicyReader{
		ordered:      ordered,
		issuanceByID: issuanceByID,
	}
}

func (r *addressPolicyReader) ListByChain(
	_ context.Context,
	chain valueobjects.SupportedChain,
) ([]entities.AddressPolicy, error) {
	policies := make([]entities.AddressPolicy, 0)
	for _, policy := range r.ordered {
		if policy.Chain != chain {
			continue
		}
		policies = append(policies, policy)
	}
	return policies, nil
}

func (r *addressPolicyReader) FindIssuanceByID(
	_ context.Context,
	addressPolicyID string,
) (entities.AddressIssuancePolicy, bool, error) {
	policy, ok := r.issuanceByID[strings.TrimSpace(addressPolicyID)]
	if !ok {
		return entities.AddressIssuancePolicy{}, false, nil
	}
	return policy, true, nil
}

func fingerprintAccountPublicKeySHA256Trunc64HexV1(accountPublicKey string) string {
	trimmed := strings.TrimSpace(accountPublicKey)
	if trimmed == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(trimmed))
	return hex.EncodeToString(sum[:8])
}
