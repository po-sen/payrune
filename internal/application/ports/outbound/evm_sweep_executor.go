package outbound

import (
	"context"
	"errors"
	"strings"

	"payrune/internal/domain/valueobjects"
)

type ExecuteEVMSweepBatchInput struct {
	Network           valueobjects.NetworkID
	RPCURL            string
	SweeperPrivateKey string
	FactoryAddress    string
	AssetType         string
	TokenAddress      string
	SaltHexes         []string
}

type ExecuteEVMSweepBatchOutput struct {
	TxHash string
}

type WaitForEVMSweepTransactionInput struct {
	Network valueobjects.NetworkID
	RPCURL  string
	TxHash  string
}

func (input ExecuteEVMSweepBatchInput) Normalize() ExecuteEVMSweepBatchInput {
	input.Network = valueobjects.NetworkID(strings.TrimSpace(string(input.Network)))
	input.RPCURL = strings.TrimSpace(input.RPCURL)
	input.SweeperPrivateKey = strings.TrimSpace(input.SweeperPrivateKey)
	input.FactoryAddress = strings.TrimSpace(input.FactoryAddress)
	input.AssetType = strings.ToLower(strings.TrimSpace(input.AssetType))
	input.TokenAddress = strings.TrimSpace(input.TokenAddress)
	normalizedSalts := make([]string, 0, len(input.SaltHexes))
	for _, saltHex := range input.SaltHexes {
		trimmed := strings.TrimSpace(strings.ToLower(saltHex))
		if trimmed == "" {
			continue
		}
		normalizedSalts = append(normalizedSalts, trimmed)
	}
	input.SaltHexes = normalizedSalts
	return input
}

func (input ExecuteEVMSweepBatchInput) Validate() (ExecuteEVMSweepBatchInput, error) {
	normalized := input.Normalize()
	if normalized.Network == "" {
		return ExecuteEVMSweepBatchInput{}, errors.New("network is required")
	}
	if normalized.RPCURL == "" {
		return ExecuteEVMSweepBatchInput{}, errors.New("rpc url is required")
	}
	if normalized.SweeperPrivateKey == "" {
		return ExecuteEVMSweepBatchInput{}, errors.New("sweeper private key is required")
	}
	if normalized.FactoryAddress == "" {
		return ExecuteEVMSweepBatchInput{}, errors.New("factory address is required")
	}
	if normalized.AssetType != "native" && normalized.AssetType != "erc20" {
		return ExecuteEVMSweepBatchInput{}, errors.New("asset type is invalid")
	}
	if normalized.AssetType == "erc20" && normalized.TokenAddress == "" {
		return ExecuteEVMSweepBatchInput{}, errors.New("token address is required")
	}
	if len(normalized.SaltHexes) == 0 {
		return ExecuteEVMSweepBatchInput{}, errors.New("salt hexes are required")
	}
	return normalized, nil
}

func (input WaitForEVMSweepTransactionInput) Validate() (WaitForEVMSweepTransactionInput, error) {
	normalized := WaitForEVMSweepTransactionInput{
		Network: valueobjects.NetworkID(strings.TrimSpace(string(input.Network))),
		RPCURL:  strings.TrimSpace(input.RPCURL),
		TxHash:  strings.TrimSpace(input.TxHash),
	}
	if normalized.Network == "" {
		return WaitForEVMSweepTransactionInput{}, errors.New("network is required")
	}
	if normalized.RPCURL == "" {
		return WaitForEVMSweepTransactionInput{}, errors.New("rpc url is required")
	}
	if normalized.TxHash == "" {
		return WaitForEVMSweepTransactionInput{}, errors.New("tx hash is required")
	}
	return normalized, nil
}

type EVMSweepExecutor interface {
	ExecuteBatch(ctx context.Context, input ExecuteEVMSweepBatchInput) (ExecuteEVMSweepBatchOutput, error)
	WaitForTransaction(ctx context.Context, input WaitForEVMSweepTransactionInput) error
}
