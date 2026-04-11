package ethereum

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

var erc20DecimalsSelector = []byte{0x31, 0x3c, 0xe5, 0x67}

func (c *ethereumRPCClient) fetchCode(
	ctx context.Context,
	address string,
) (string, error) {
	var rawCode string
	if err := c.call(ctx, "eth_getCode", []any{address, "latest"}, &rawCode); err != nil {
		return "", err
	}
	if _, err := parseEthereumHexBytes(rawCode, "runtime code"); err != nil {
		return "", err
	}
	return strings.ToLower(strings.TrimSpace(rawCode)), nil
}

func (c *ethereumRPCClient) fetchERC20Decimals(
	ctx context.Context,
	erc20AssetReference string,
) (uint8, error) {
	rawDecimals, err := c.callContractAtLatest(
		ctx,
		erc20AssetReference,
		encodeERC20DecimalsCall(),
	)
	if err != nil {
		return 0, err
	}
	return parseEthereumHexQuantityToUint8(rawDecimals, "token decimals")
}

func (c *ethereumRPCClient) callContractAtLatest(
	ctx context.Context,
	address string,
	callData string,
) (string, error) {
	return c.callContract(ctx, address, callData, "latest")
}

func (c *ethereumRPCClient) callContract(
	ctx context.Context,
	address string,
	callData string,
	blockTag string,
) (string, error) {
	var rawResult string
	if err := c.call(
		ctx,
		"eth_call",
		[]any{
			map[string]string{
				"to":   address,
				"data": callData,
			},
			blockTag,
		},
		&rawResult,
	); err != nil {
		return "", err
	}
	return rawResult, nil
}

func parseEthereumHexQuantityToUint8(raw string, label string) (uint8, error) {
	value, err := parseEthereumHexQuantity(raw, label)
	if err != nil {
		return 0, err
	}
	if value.Sign() < 0 {
		return 0, fmt.Errorf("%s must be non-negative", label)
	}
	if !value.IsUint64() || value.Uint64() > uint64(^uint8(0)) {
		return 0, fmt.Errorf("%s exceeds uint8", label)
	}
	return uint8(value.Uint64()), nil
}

func parseEthereumHexQuantity(raw string, label string) (*big.Int, error) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "0x") && !strings.HasPrefix(trimmed, "0X") {
		return nil, fmt.Errorf("%s must start with 0x", label)
	}

	value, ok := new(big.Int).SetString(trimmed[2:], 16)
	if !ok {
		return nil, fmt.Errorf("%s is invalid hex", label)
	}
	return value, nil
}

func parseEthereumHexBytes(raw string, label string) ([]byte, error) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "0x") && !strings.HasPrefix(trimmed, "0X") {
		return nil, fmt.Errorf("%s must start with 0x", label)
	}
	if len(trimmed)%2 != 0 {
		return nil, fmt.Errorf("%s must have even-length hex", label)
	}

	decoded, err := hex.DecodeString(trimmed[2:])
	if err != nil {
		return nil, fmt.Errorf("%s is invalid hex: %w", label, err)
	}
	return decoded, nil
}

func encodeERC20DecimalsCall() string {
	return "0x" + hex.EncodeToString(erc20DecimalsSelector)
}
