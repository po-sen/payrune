package policy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	outport "payrune/internal/application/ports/out"
	"payrune/internal/domain/entities"
	"payrune/internal/domain/value_objects"
)

const xpubFingerprintAlgorithmSHA256Trunc64HexV1 = "sha256-trunc64-hex-v1"

type AddressPolicyConfig struct {
	AddressPolicyID      string
	Chain                value_objects.Chain
	Network              value_objects.BitcoinNetwork
	Scheme               value_objects.BitcoinAddressScheme
	MinorUnit            string
	Decimals             uint8
	XPub                 string
	XPubFingerprintAlgo  string
	XPubFingerprint      string
	DerivationPathPrefix string
}

type addressPolicyReader struct {
	ordered []entities.AddressPolicy
	byID    map[string]entities.AddressPolicy
}

var _ outport.AddressPolicyReader = (*addressPolicyReader)(nil)

func NewAddressPolicyReader(configs []AddressPolicyConfig) outport.AddressPolicyReader {
	ordered := make([]entities.AddressPolicy, 0, len(configs))
	byID := make(map[string]entities.AddressPolicy, len(configs))

	for _, cfg := range configs {
		normalized := entities.AddressPolicy{
			AddressPolicyID:      strings.TrimSpace(cfg.AddressPolicyID),
			Chain:                cfg.Chain,
			Network:              cfg.Network,
			Scheme:               cfg.Scheme,
			MinorUnit:            strings.TrimSpace(cfg.MinorUnit),
			Decimals:             cfg.Decimals,
			XPub:                 strings.TrimSpace(cfg.XPub),
			XPubFingerprintAlgo:  strings.TrimSpace(cfg.XPubFingerprintAlgo),
			XPubFingerprint:      strings.TrimSpace(cfg.XPubFingerprint),
			DerivationPathPrefix: strings.TrimSpace(cfg.DerivationPathPrefix),
		}.Normalize()

		if normalized.XPub != "" && normalized.XPubFingerprintAlgo == "" {
			normalized.XPubFingerprintAlgo = xpubFingerprintAlgorithmSHA256Trunc64HexV1
		}
		if normalized.XPub != "" && normalized.XPubFingerprint == "" {
			normalized.XPubFingerprintAlgo = xpubFingerprintAlgorithmSHA256Trunc64HexV1
			normalized.XPubFingerprint = fingerprintSHA256Trunc64HexV1(normalized.XPub)
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

	return &addressPolicyReader{
		ordered: ordered,
		byID:    byID,
	}
}

func (r *addressPolicyReader) ListByChain(
	_ context.Context,
	chain value_objects.Chain,
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

func (r *addressPolicyReader) FindByID(
	_ context.Context,
	addressPolicyID string,
) (entities.AddressPolicy, bool, error) {
	policy, ok := r.byID[strings.TrimSpace(addressPolicyID)]
	if !ok {
		return entities.AddressPolicy{}, false, nil
	}
	return policy, true, nil
}

func fingerprintSHA256Trunc64HexV1(xpub string) string {
	trimmed := strings.TrimSpace(xpub)
	if trimmed == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(trimmed))
	return hex.EncodeToString(sum[:8])
}
