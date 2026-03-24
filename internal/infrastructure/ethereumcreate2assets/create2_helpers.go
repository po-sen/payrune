package ethereumcreate2assets

import (
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"
)

const create2SourceRefVersion = "create2.v1"

func buildCreate2AddressSourceRef(
	factoryAddress string,
	collectorAddress string,
	initCodeHash string,
) (string, error) {
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

func buildFixedCollectorReceiverInitCodeHex(creationCodeHex string, collectorAddress string) (string, error) {
	creationCodeHex = strings.TrimSpace(creationCodeHex)
	if !strings.HasPrefix(creationCodeHex, "0x") && !strings.HasPrefix(creationCodeHex, "0X") {
		return "", fmt.Errorf("receiver creation code hex is invalid")
	}
	if len(creationCodeHex) <= 2 {
		return "", fmt.Errorf("receiver creation code hex is invalid")
	}

	creationCode, err := hex.DecodeString(creationCodeHex[2:])
	if err != nil {
		return "", err
	}
	if len(creationCode) == 0 {
		return "", fmt.Errorf("receiver creation code hex is invalid")
	}

	_, collectorBytes, err := normalizeFixedHex(collectorAddress, 20, "collector address")
	if err != nil {
		return "", err
	}

	encodedCollector := make([]byte, 32)
	copy(encodedCollector[12:], collectorBytes)

	initCode := append(append([]byte{}, creationCode...), encodedCollector...)
	return "0x" + hex.EncodeToString(initCode), nil
}

func keccak256Hex(rawHex string) (string, error) {
	rawHex = strings.TrimSpace(rawHex)
	if !strings.HasPrefix(rawHex, "0x") && !strings.HasPrefix(rawHex, "0X") {
		return "", fmt.Errorf("hex value is invalid")
	}
	if len(rawHex) <= 2 {
		return "", fmt.Errorf("hex value is invalid")
	}

	decoded, err := hex.DecodeString(rawHex[2:])
	if err != nil {
		return "", err
	}
	if len(decoded) == 0 {
		return "", fmt.Errorf("hex value is invalid")
	}

	sum := keccak256Hash(decoded)
	return "0x" + hex.EncodeToString(sum[:]), nil
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

func keccak256Hash(data []byte) [32]byte {
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write(data)

	var out [32]byte
	sum := hasher.Sum(nil)
	copy(out[:], sum)
	return out
}
