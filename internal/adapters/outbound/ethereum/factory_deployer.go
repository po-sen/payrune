package ethereum

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	gethethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

const defaultFactoryContractsArtifactPath = "deployments/ethereum/build/contracts.json"

type FactoryDeployer struct{}

func NewFactoryDeployer() *FactoryDeployer {
	return &FactoryDeployer{}
}

func (d *FactoryDeployer) Deploy(
	ctx context.Context,
	input outport.DeployEVMFactoryInput,
) (outport.DeployEVMFactoryOutput, error) {
	validated, err := input.Validate()
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}
	if !common.IsHexAddress(validated.CollectorAddress) {
		return outport.DeployEVMFactoryOutput{}, errors.New("collector address is invalid")
	}

	artifact, err := loadDepositVaultFactoryArtifact(validated.ContractsArtifactPath)
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}

	client, err := ethclient.DialContext(ctx, validated.RPCURL)
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}
	if err := ensureEVMDeployNetwork(validated.Network, chainID.String()); err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}

	privateKey, err := parsePrivateKey(validated.DeployPrivateKey)
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}
	auth.Context = ctx

	contractABI, err := abi.JSON(strings.NewReader(artifact.ABIJSON))
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}
	contractAddress, deployTx, _, err := bind.DeployContract(
		auth,
		contractABI,
		artifact.Bytecode,
		client,
		common.HexToAddress(validated.CollectorAddress),
	)
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}

	receipt, err := waitForDeploymentReceipt(ctx, client, deployTx, validated.Confirmations)
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}
	if receipt.Status != types.ReceiptStatusSuccessful {
		return outport.DeployEVMFactoryOutput{}, errors.New("evm factory deployment reverted")
	}

	vaultCreationCodeHash, err := callFactoryBytes32View(
		ctx,
		client,
		contractAddress,
		contractABI,
		"vaultCreationCodeHash",
	)
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}

	manifest := outport.EVMFactoryDeploymentManifest{
		ContractName:              artifact.ContractName,
		SourceName:                artifact.SourceName,
		CompilerVersion:           artifact.CompilerVersion,
		ChainID:                   chainID.String(),
		ContractAddress:           contractAddress.Hex(),
		Collector:                 common.HexToAddress(validated.CollectorAddress).Hex(),
		VaultCreationCodeHash:     vaultCreationCodeHash,
		Deployer:                  auth.From.Hex(),
		DeploymentTransactionHash: deployTx.Hash().Hex(),
		Confirmations:             validated.Confirmations,
		DeployedAt:                time.Now().UTC(),
	}
	manifest, err = manifest.Validate()
	if err != nil {
		return outport.DeployEVMFactoryOutput{}, err
	}

	output := outport.DeployEVMFactoryOutput{
		Manifest:           manifest,
		OutputManifestPath: validated.OutputManifestPath,
	}
	if validated.OutputManifestPath != "" {
		if err := writeEVMFactoryDeploymentManifest(validated.OutputManifestPath, manifest); err != nil {
			return outport.DeployEVMFactoryOutput{}, err
		}
	}

	return output, nil
}

type depositVaultFactoryArtifact struct {
	ContractName    string
	SourceName      string
	CompilerVersion string
	ABIJSON         string
	Bytecode        []byte
}

type contractsBuildFile struct {
	CompilerVersion string                                        `json:"compilerVersion"`
	Contracts       map[string]map[string]contractsBuildFileEntry `json:"contracts"`
}

type contractsBuildFileEntry struct {
	ABI []any `json:"abi"`
	EVM struct {
		Bytecode struct {
			Object string `json:"object"`
		} `json:"bytecode"`
	} `json:"evm"`
}

func loadDepositVaultFactoryArtifact(path string) (depositVaultFactoryArtifact, error) {
	artifactPath := strings.TrimSpace(path)
	if artifactPath == "" {
		artifactPath = defaultFactoryContractsArtifactPath
	}
	content, err := os.ReadFile(filepath.Clean(artifactPath))
	if err != nil {
		return depositVaultFactoryArtifact{}, fmt.Errorf("contracts artifact could not be read: %w", err)
	}

	var buildFile contractsBuildFile
	if err := json.Unmarshal(content, &buildFile); err != nil {
		return depositVaultFactoryArtifact{}, fmt.Errorf("contracts artifact is not valid json: %w", err)
	}

	sourceContracts, ok := buildFile.Contracts["DepositVaultFactory.sol"]
	if !ok {
		return depositVaultFactoryArtifact{}, errors.New("contracts artifact is missing DepositVaultFactory.sol")
	}
	entry, ok := sourceContracts["DepositVaultFactory"]
	if !ok {
		return depositVaultFactoryArtifact{}, errors.New("contracts artifact is missing DepositVaultFactory")
	}

	abiJSON, err := json.Marshal(entry.ABI)
	if err != nil {
		return depositVaultFactoryArtifact{}, fmt.Errorf("contracts artifact abi is invalid: %w", err)
	}
	bytecode, err := hex.DecodeString(strings.TrimSpace(entry.EVM.Bytecode.Object))
	if err != nil {
		return depositVaultFactoryArtifact{}, fmt.Errorf("contracts artifact bytecode is invalid: %w", err)
	}
	if len(bytecode) == 0 {
		return depositVaultFactoryArtifact{}, errors.New("contracts artifact bytecode is empty")
	}

	return depositVaultFactoryArtifact{
		ContractName:    "DepositVaultFactory",
		SourceName:      "DepositVaultFactory.sol",
		CompilerVersion: strings.TrimSpace(buildFile.CompilerVersion),
		ABIJSON:         string(abiJSON),
		Bytecode:        bytecode,
	}, nil
}

func ensureEVMDeployNetwork(network valueobjects.NetworkID, chainID string) error {
	switch strings.TrimSpace(string(network)) {
	case "mainnet":
		if chainID != "1" {
			return fmt.Errorf("ethereum network mismatch: expected chainId 1 for mainnet, got %s", chainID)
		}
	case "sepolia":
		if chainID != "11155111" {
			return fmt.Errorf("ethereum network mismatch: expected chainId 11155111 for sepolia, got %s", chainID)
		}
	default:
		return fmt.Errorf("unsupported ethereum deploy network: %s", network)
	}
	return nil
}

func waitForDeploymentReceipt(
	ctx context.Context,
	client *ethclient.Client,
	tx *types.Transaction,
	confirmations int,
) (*types.Receipt, error) {
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		return nil, err
	}
	if confirmations <= 1 {
		return receipt, nil
	}

	targetBlock := receipt.BlockNumber.Uint64() + uint64(confirmations-1)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		header, err := client.HeaderByNumber(ctx, nil)
		if err != nil {
			return nil, err
		}
		if header.Number.Uint64() >= targetBlock {
			return receipt, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func callFactoryBytes32View(
	ctx context.Context,
	client *ethclient.Client,
	contractAddress common.Address,
	contractABI abi.ABI,
	method string,
) (string, error) {
	callData, err := contractABI.Pack(method)
	if err != nil {
		return "", err
	}
	rawResult, err := client.CallContract(
		ctx,
		gethethereum.CallMsg{
			To:   &contractAddress,
			Data: callData,
		},
		nil,
	)
	if err != nil {
		return "", err
	}
	results, err := contractABI.Unpack(method, rawResult)
	if err != nil {
		return "", err
	}
	if len(results) != 1 {
		return "", errors.New("unexpected view return value count")
	}

	switch value := results[0].(type) {
	case [32]byte:
		return "0x" + hex.EncodeToString(value[:]), nil
	case common.Hash:
		return value.Hex(), nil
	default:
		return "", fmt.Errorf("unexpected %s return type: %T", method, results[0])
	}
}

func writeEVMFactoryDeploymentManifest(
	path string,
	manifest outport.EVMFactoryDeploymentManifest,
) error {
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("deployment manifest could not be encoded: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("deployment manifest directory could not be created: %w", err)
	}
	if err := os.WriteFile(filepath.Clean(path), content, 0o600); err != nil {
		return fmt.Errorf("deployment manifest could not be written: %w", err)
	}
	return nil
}

var _ outport.EVMFactoryDeployer = (*FactoryDeployer)(nil)
