package ethereum

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/sha3"
)

func GenerateSaltHex(reader io.Reader) (string, error) {
	if reader == nil {
		return "", errors.New("random reader is required")
	}

	salt := make([]byte, 32)
	if _, err := io.ReadFull(reader, salt); err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(salt), nil
}

func PredictVaultAddress(factoryAddress string, saltHex string, vaultCreationCodeHash string) (string, error) {
	factoryBytes, err := decodeHexBytes(factoryAddress, 20)
	if err != nil {
		return "", fmt.Errorf("factory address is invalid: %w", err)
	}
	saltBytes, err := decodeHexBytes(saltHex, 32)
	if err != nil {
		return "", fmt.Errorf("salt hex is invalid: %w", err)
	}
	hashBytes, err := decodeHexBytes(vaultCreationCodeHash, 32)
	if err != nil {
		return "", fmt.Errorf("vault creation code hash is invalid: %w", err)
	}

	payload := make([]byte, 0, 1+20+32+32)
	payload = append(payload, 0xff)
	payload = append(payload, factoryBytes...)
	payload = append(payload, saltBytes...)
	payload = append(payload, hashBytes...)

	sum := keccak256(payload)
	return "0x" + hex.EncodeToString(sum[12:]), nil
}

func decodeHexBytes(value string, size int) ([]byte, error) {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "0x") {
		return nil, errors.New("missing 0x prefix")
	}
	raw := trimmed[2:]
	if len(raw) != size*2 {
		return nil, fmt.Errorf("expected %d-byte hex", size)
	}
	bytes, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func keccak256(payload []byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	_, _ = hasher.Write(payload)
	return hasher.Sum(nil)
}
