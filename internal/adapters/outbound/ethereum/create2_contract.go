package ethereum

import (
	"encoding/hex"
	"errors"
	"strings"
)

func BuildFixedCollectorReceiverInitCodeHex(creationCodeHex string, collectorAddress string) (string, error) {
	creationCodeHex = strings.TrimSpace(creationCodeHex)
	if !strings.HasPrefix(creationCodeHex, "0x") && !strings.HasPrefix(creationCodeHex, "0X") {
		return "", errors.New("receiver creation code hex is invalid")
	}
	if len(creationCodeHex) <= 2 {
		return "", errors.New("receiver creation code hex is invalid")
	}

	creationCode, err := hex.DecodeString(creationCodeHex[2:])
	if err != nil {
		return "", err
	}
	if len(creationCode) == 0 {
		return "", errors.New("receiver creation code hex is invalid")
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

func Keccak256Hex(rawHex string) (string, error) {
	rawHex = strings.TrimSpace(rawHex)
	if !strings.HasPrefix(rawHex, "0x") && !strings.HasPrefix(rawHex, "0X") {
		return "", errors.New("hex value is invalid")
	}
	if len(rawHex) <= 2 {
		return "", errors.New("hex value is invalid")
	}

	decoded, err := hex.DecodeString(rawHex[2:])
	if err != nil {
		return "", err
	}
	if len(decoded) == 0 {
		return "", errors.New("hex value is invalid")
	}

	sum := keccak256Hash(decoded)
	return "0x" + hex.EncodeToString(sum[:]), nil
}
