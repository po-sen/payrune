package outbound

import (
	"context"
	"errors"
	"strings"
	"time"

	"payrune/internal/domain/valueobjects"
)

type DeployEVMFactoryInput struct {
	Network               valueobjects.NetworkID
	RPCURL                string
	DeployPrivateKey      string
	CollectorAddress      string
	Confirmations         int
	OutputManifestPath    string
	ContractsArtifactPath string
}

func (input DeployEVMFactoryInput) Normalize() DeployEVMFactoryInput {
	input.Network = valueobjects.NetworkID(strings.TrimSpace(string(input.Network)))
	input.RPCURL = strings.TrimSpace(input.RPCURL)
	input.DeployPrivateKey = strings.TrimSpace(input.DeployPrivateKey)
	input.CollectorAddress = strings.TrimSpace(input.CollectorAddress)
	input.OutputManifestPath = strings.TrimSpace(input.OutputManifestPath)
	input.ContractsArtifactPath = strings.TrimSpace(input.ContractsArtifactPath)
	return input
}

func (input DeployEVMFactoryInput) Validate() (DeployEVMFactoryInput, error) {
	normalized := input.Normalize()
	if normalized.Network == "" {
		return DeployEVMFactoryInput{}, errors.New("network is required")
	}
	if normalized.RPCURL == "" {
		return DeployEVMFactoryInput{}, errors.New("rpc url is required")
	}
	if normalized.DeployPrivateKey == "" {
		return DeployEVMFactoryInput{}, errors.New("deploy private key is required")
	}
	if normalized.CollectorAddress == "" {
		return DeployEVMFactoryInput{}, errors.New("collector address is required")
	}
	if normalized.Confirmations < 0 {
		return DeployEVMFactoryInput{}, errors.New("confirmations must be greater than or equal to zero")
	}
	return normalized, nil
}

type EVMFactoryDeploymentManifest struct {
	ContractName              string    `json:"contractName"`
	SourceName                string    `json:"sourceName,omitempty"`
	CompilerVersion           string    `json:"compilerVersion,omitempty"`
	ChainID                   string    `json:"chainId"`
	ContractAddress           string    `json:"contractAddress"`
	Collector                 string    `json:"collector"`
	VaultCreationCodeHash     string    `json:"vaultCreationCodeHash"`
	Deployer                  string    `json:"deployer,omitempty"`
	DeploymentTransactionHash string    `json:"deploymentTransactionHash,omitempty"`
	Confirmations             int       `json:"confirmations,omitempty"`
	DeployedAt                time.Time `json:"deployedAt"`
}

func (manifest EVMFactoryDeploymentManifest) Normalize() EVMFactoryDeploymentManifest {
	manifest.ContractName = strings.TrimSpace(manifest.ContractName)
	manifest.SourceName = strings.TrimSpace(manifest.SourceName)
	manifest.CompilerVersion = strings.TrimSpace(manifest.CompilerVersion)
	manifest.ChainID = strings.TrimSpace(manifest.ChainID)
	manifest.ContractAddress = strings.TrimSpace(manifest.ContractAddress)
	manifest.Collector = strings.TrimSpace(manifest.Collector)
	manifest.VaultCreationCodeHash = strings.TrimSpace(manifest.VaultCreationCodeHash)
	manifest.Deployer = strings.TrimSpace(manifest.Deployer)
	manifest.DeploymentTransactionHash = strings.TrimSpace(manifest.DeploymentTransactionHash)
	if !manifest.DeployedAt.IsZero() {
		manifest.DeployedAt = manifest.DeployedAt.UTC()
	}
	return manifest
}

func (manifest EVMFactoryDeploymentManifest) Validate() (EVMFactoryDeploymentManifest, error) {
	normalized := manifest.Normalize()
	if normalized.ChainID == "" {
		return EVMFactoryDeploymentManifest{}, errors.New("chain id is required")
	}
	if normalized.ContractAddress == "" {
		return EVMFactoryDeploymentManifest{}, errors.New("contract address is required")
	}
	if normalized.Collector == "" {
		return EVMFactoryDeploymentManifest{}, errors.New("collector is required")
	}
	if normalized.VaultCreationCodeHash == "" {
		return EVMFactoryDeploymentManifest{}, errors.New("vault creation code hash is required")
	}
	if normalized.DeployedAt.IsZero() {
		return EVMFactoryDeploymentManifest{}, errors.New("deployed at is required")
	}
	return normalized, nil
}

type DeployEVMFactoryOutput struct {
	Manifest           EVMFactoryDeploymentManifest
	OutputManifestPath string
}

type EVMFactoryDeployer interface {
	Deploy(ctx context.Context, input DeployEVMFactoryInput) (DeployEVMFactoryOutput, error)
}
