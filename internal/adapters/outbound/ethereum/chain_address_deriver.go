package ethereum

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

const create2SourceRefVersion = "create2.v1"

type create2SourceRef struct {
	factoryAddress string
	collector      string
	initCodeHash   string
}

type ChainAddressDeriver struct{}

func NewChainAddressDeriver() *ChainAddressDeriver {
	return &ChainAddressDeriver{}
}

func BuildCreate2AddressSpaceRef(
	factoryAddress string,
	collectorAddress string,
	initCodeHash string,
) (string, error) {
	// Source-ref material must stay limited to fixed address-space inputs.
	// The gas-paying operator signer is intentionally excluded so signer
	// rotation does not change predicted payment addresses.
	factoryAddress, _, err := normalizeFixedHex(factoryAddress, 20, "factory address")
	if err != nil {
		return "", err
	}
	collectorAddress, _, err = normalizeFixedHex(collectorAddress, 20, "collector address")
	if err != nil {
		return "", err
	}
	initCodeHash, _, err = normalizeFixedHex(initCodeHash, 32, "init code hash")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"%s:factory=%s;collector=%s;init_code_hash=%s",
		create2SourceRefVersion,
		factoryAddress,
		collectorAddress,
		initCodeHash,
	), nil
}

func (g *ChainAddressDeriver) Chain() valueobjects.SupportedChain {
	return valueobjects.SupportedChainEthereum
}

func (g *ChainAddressDeriver) DeriveAddress(
	_ context.Context,
	input outport.DeriveChainAddressInput,
) (outport.DeriveChainAddressOutput, error) {
	if input.Chain != valueobjects.SupportedChainEthereum {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}
	if !input.Scheme.Normalize().IsCreate2() {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}

	sourceRef, err := parseCreate2SourceRef(input.AddressSpaceRef)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}

	factoryAddress, err := mustDecodeFixedHex(sourceRef.factoryAddress, 20)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}
	initCodeHash, err := mustDecodeFixedHex(sourceRef.initCodeHash, 32)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}

	normalizedSaltHex, salt, err := normalizeCreate2Salt(input.RelativeIssuanceRef)
	if err != nil {
		return outport.DeriveChainAddressOutput{}, outport.ErrChainAddressDerivationInputInvalid
	}
	predictedAddress := predictCreate2Address(factoryAddress, salt, initCodeHash)

	return outport.DeriveChainAddressOutput{
		Address:             predictedAddress,
		RelativeIssuanceRef: normalizedSaltHex,
		IssuanceRefKind:     valueobjects.IssuanceRefKindCreate2Salt,
		IssuanceRef:         normalizedSaltHex,
	}, nil
}

func parseCreate2SourceRef(raw string) (create2SourceRef, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return create2SourceRef{}, errors.New("ethereum address source ref is required")
	}
	if !strings.HasPrefix(trimmed, create2SourceRefVersion+":") {
		return create2SourceRef{}, errors.New("ethereum address source ref format is invalid")
	}

	payload := strings.TrimPrefix(trimmed, create2SourceRefVersion+":")
	parts := strings.Split(payload, ";")
	if len(parts) != 3 {
		return create2SourceRef{}, errors.New("ethereum address source ref format is invalid")
	}

	values := make(map[string]string, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return create2SourceRef{}, errors.New("ethereum address source ref format is invalid")
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return create2SourceRef{}, errors.New("ethereum address source ref format is invalid")
		}
		if _, exists := values[key]; exists {
			return create2SourceRef{}, errors.New("ethereum address source ref format is invalid")
		}
		values[key] = value
	}

	factoryAddress, _, err := normalizeFixedHex(values["factory"], 20, "factory address")
	if err != nil {
		return create2SourceRef{}, err
	}
	collectorAddress, _, err := normalizeFixedHex(values["collector"], 20, "collector address")
	if err != nil {
		return create2SourceRef{}, err
	}
	initCodeHash, _, err := normalizeFixedHex(values["init_code_hash"], 32, "init code hash")
	if err != nil {
		return create2SourceRef{}, err
	}

	return create2SourceRef{
		factoryAddress: factoryAddress,
		collector:      collectorAddress,
		initCodeHash:   initCodeHash,
	}, nil
}

func normalizeCreate2Salt(raw string) (string, [32]byte, error) {
	normalizedSalt, saltBytes, err := normalizeFixedHex(raw, 32, "ethereum relative address reference")
	if err != nil {
		return "", [32]byte{}, err
	}

	var salt [32]byte
	copy(salt[:], saltBytes)
	return normalizedSalt, salt, nil
}

func predictCreate2Address(factoryAddress []byte, salt [32]byte, initCodeHash []byte) string {
	preimage := make([]byte, 0, 1+len(factoryAddress)+len(salt)+len(initCodeHash))
	preimage = append(preimage, 0xff)
	preimage = append(preimage, factoryAddress...)
	preimage = append(preimage, salt[:]...)
	preimage = append(preimage, initCodeHash...)
	digest := keccak256Hash(preimage)
	return "0x" + hex.EncodeToString(digest[12:])
}

func normalizeFixedHex(raw string, sizeBytes int, label string) (string, []byte, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil, fmt.Errorf("%s is required", label)
	}
	if !strings.HasPrefix(trimmed, "0x") && !strings.HasPrefix(trimmed, "0X") {
		return "", nil, fmt.Errorf("%s must start with 0x", label)
	}

	decoded, err := hex.DecodeString(trimmed[2:])
	if err != nil {
		return "", nil, fmt.Errorf("%s is invalid hex: %w", label, err)
	}
	if len(decoded) != sizeBytes {
		return "", nil, fmt.Errorf("%s must be %d bytes", label, sizeBytes)
	}

	return "0x" + strings.ToLower(trimmed[2:]), decoded, nil
}

func mustDecodeFixedHex(raw string, sizeBytes int) ([]byte, error) {
	_, decoded, err := normalizeFixedHex(raw, sizeBytes, "hex value")
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func keccak256Hash(data []byte) [32]byte {
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write(data)

	var out [32]byte
	sum := hasher.Sum(nil)
	copy(out[:], sum)
	return out
}
