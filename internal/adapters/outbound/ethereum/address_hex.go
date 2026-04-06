package ethereum

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func NormalizeEthereumAddress(raw string, label string) (string, error) {
	normalized, _, err := normalizeFixedHex(raw, 20, label)
	if err != nil {
		return "", err
	}
	return normalized, nil
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
