package ethereum

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const create2SaltDerivationVersion = "ethereum-create2-salt.v1"

type Create2SaltDeriver struct {
	keyByNetwork map[string][]byte
}

type DeriveCreate2AllocationSaltInput struct {
	Network          string
	AddressPolicyID  string
	PaymentAddressID int64
	SlotIndex        uint32
}

func NewCreate2SaltDeriver(rawSecretsByNetwork map[string]string) *Create2SaltDeriver {
	keyByNetwork := make(map[string][]byte, len(rawSecretsByNetwork))
	for network, rawSecret := range rawSecretsByNetwork {
		_, secretKey, ok := normalizeCreate2SaltSecret(rawSecret)
		if !ok {
			continue
		}
		keyByNetwork[strings.ToLower(strings.TrimSpace(network))] = secretKey
	}

	return &Create2SaltDeriver{keyByNetwork: keyByNetwork}
}

func (d *Create2SaltDeriver) HasNetwork(network string) bool {
	if d == nil {
		return false
	}
	_, ok := d.keyByNetwork[strings.ToLower(strings.TrimSpace(network))]
	return ok
}

func (d *Create2SaltDeriver) DeriveAllocationSalt(
	_ context.Context,
	input DeriveCreate2AllocationSaltInput,
) (string, error) {
	if d == nil {
		return "", errors.New("ethereum create2 salt deriver is not configured")
	}

	network := strings.ToLower(strings.TrimSpace(input.Network))
	secretKey, ok := d.keyByNetwork[network]
	if !ok {
		return "", fmt.Errorf("ethereum create2 salt deriver is not configured for network: %s", input.Network)
	}
	addressPolicyID := strings.ToLower(strings.TrimSpace(input.AddressPolicyID))
	if addressPolicyID == "" {
		return "", errors.New("address policy id is required")
	}
	if input.PaymentAddressID <= 0 {
		return "", errors.New("payment address id must be greater than zero")
	}

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(create2SaltDerivationVersion))
	mac.Write([]byte{'\n'})
	mac.Write([]byte(network))
	mac.Write([]byte{'\n'})
	mac.Write([]byte(addressPolicyID))
	mac.Write([]byte{'\n'})
	mac.Write([]byte(strconv.FormatInt(input.PaymentAddressID, 10)))
	mac.Write([]byte{'\n'})
	mac.Write([]byte(strconv.FormatUint(uint64(input.SlotIndex), 10)))
	sum := mac.Sum(nil)
	return "0x" + hex.EncodeToString(sum), nil
}

func normalizeCreate2SaltSecret(raw string) (string, []byte, bool) {
	normalized, decoded, err := normalizeFixedHex(raw, 32, "ethereum create2 derivation key")
	if err != nil {
		return "", nil, false
	}
	return normalized, decoded, true
}
