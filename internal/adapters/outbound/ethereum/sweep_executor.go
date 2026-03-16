package ethereum

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	gethethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	outport "payrune/internal/application/ports/outbound"
)

const depositVaultFactoryABIJSON = `[
  {"inputs":[{"internalType":"bytes32[]","name":"salts","type":"bytes32[]"}],"name":"batchDeployAndSweepNative","outputs":[],"stateMutability":"nonpayable","type":"function"},
  {"inputs":[{"internalType":"bytes32[]","name":"salts","type":"bytes32[]"},{"internalType":"address","name":"token","type":"address"}],"name":"batchDeployAndSweepToken","outputs":[],"stateMutability":"nonpayable","type":"function"}
]`

type SweepExecutor struct{}

func NewSweepExecutor() *SweepExecutor {
	return &SweepExecutor{}
}

func (e *SweepExecutor) ExecuteBatch(
	ctx context.Context,
	input outport.ExecuteEVMSweepBatchInput,
) (outport.ExecuteEVMSweepBatchOutput, error) {
	validated, err := input.Validate()
	if err != nil {
		return outport.ExecuteEVMSweepBatchOutput{}, err
	}
	if !common.IsHexAddress(validated.FactoryAddress) {
		return outport.ExecuteEVMSweepBatchOutput{}, errors.New("factory address is invalid")
	}
	if validated.AssetType == "erc20" && !common.IsHexAddress(validated.TokenAddress) {
		return outport.ExecuteEVMSweepBatchOutput{}, errors.New("token address is invalid")
	}

	client, err := ethclient.DialContext(ctx, validated.RPCURL)
	if err != nil {
		return outport.ExecuteEVMSweepBatchOutput{}, err
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return outport.ExecuteEVMSweepBatchOutput{}, err
	}

	privateKey, err := parsePrivateKey(validated.SweeperPrivateKey)
	if err != nil {
		return outport.ExecuteEVMSweepBatchOutput{}, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return outport.ExecuteEVMSweepBatchOutput{}, err
	}
	auth.Context = ctx

	contractABI, err := abi.JSON(strings.NewReader(depositVaultFactoryABIJSON))
	if err != nil {
		return outport.ExecuteEVMSweepBatchOutput{}, err
	}
	contract := bind.NewBoundContract(
		common.HexToAddress(validated.FactoryAddress),
		contractABI,
		client,
		client,
		client,
	)

	salts, err := decodeSaltHexes(validated.SaltHexes)
	if err != nil {
		return outport.ExecuteEVMSweepBatchOutput{}, err
	}

	var tx *types.Transaction
	switch validated.AssetType {
	case "native":
		tx, err = contract.Transact(auth, "batchDeployAndSweepNative", salts)
	case "erc20":
		tx, err = contract.Transact(auth, "batchDeployAndSweepToken", salts, common.HexToAddress(validated.TokenAddress))
	default:
		err = fmt.Errorf("unsupported asset type: %s", validated.AssetType)
	}
	if err != nil {
		return outport.ExecuteEVMSweepBatchOutput{}, err
	}

	return outport.ExecuteEVMSweepBatchOutput{TxHash: tx.Hash().Hex()}, nil
}

func (e *SweepExecutor) WaitForTransaction(
	ctx context.Context,
	input outport.WaitForEVMSweepTransactionInput,
) error {
	validated, err := input.Validate()
	if err != nil {
		return err
	}
	if !common.IsHexHash(validated.TxHash) {
		return errors.New("tx hash is invalid")
	}

	client, err := ethclient.DialContext(ctx, validated.RPCURL)
	if err != nil {
		return err
	}
	defer client.Close()

	txHash := common.HexToHash(validated.TxHash)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == nil {
			if receipt.Status != types.ReceiptStatusSuccessful {
				return errors.New("evm sweep transaction reverted")
			}
			return nil
		}
		if !errors.Is(err, gethethereum.NotFound) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func parsePrivateKey(raw string) (*ecdsa.PrivateKey, error) {
	trimmed := strings.TrimPrefix(strings.TrimSpace(raw), "0x")
	privateKey, err := crypto.HexToECDSA(trimmed)
	if err != nil {
		return nil, errors.New("sweeper private key is invalid")
	}
	return privateKey, nil
}

func decodeSaltHexes(values []string) ([][32]byte, error) {
	decoded := make([][32]byte, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimPrefix(strings.TrimSpace(strings.ToLower(value)), "0x")
		if len(trimmed) != 64 {
			return nil, fmt.Errorf("salt hex is invalid: %s", value)
		}
		raw, err := hex.DecodeString(trimmed)
		if err != nil {
			return nil, fmt.Errorf("salt hex is invalid: %w", err)
		}
		var salt [32]byte
		copy(salt[:], raw)
		decoded = append(decoded, salt)
	}
	return decoded, nil
}

var _ outport.EVMSweepExecutor = (*SweepExecutor)(nil)
