package di

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
)

const (
	envEthereumMainnetRPCURL              = "ETHEREUM_MAINNET_RPC_URL"
	envEthereumMainnetDeployPrivateKey    = "ETHEREUM_MAINNET_DEPLOY_PRIVATE_KEY"
	envEthereumMainnetDeployConfirmations = "ETHEREUM_MAINNET_DEPLOY_CONFIRMATIONS"
	envEthereumMainnetSweeperPrivateKey   = "ETHEREUM_MAINNET_SWEEPER_PRIVATE_KEY"
	envEthereumMainnetCollectorAddress    = "ETHEREUM_MAINNET_COLLECTOR_ADDRESS"
	envEthereumMainnetFactoryManifest     = "ETHEREUM_MAINNET_FACTORY_DEPLOYMENT_MANIFEST"

	envEthereumSepoliaRPCURL              = "ETHEREUM_SEPOLIA_RPC_URL"
	envEthereumSepoliaDeployPrivateKey    = "ETHEREUM_SEPOLIA_DEPLOY_PRIVATE_KEY"
	envEthereumSepoliaDeployConfirmations = "ETHEREUM_SEPOLIA_DEPLOY_CONFIRMATIONS"
	envEthereumSepoliaSweeperPrivateKey   = "ETHEREUM_SEPOLIA_SWEEPER_PRIVATE_KEY"
	envEthereumSepoliaCollectorAddress    = "ETHEREUM_SEPOLIA_COLLECTOR_ADDRESS"
	envEthereumSepoliaFactoryManifest     = "ETHEREUM_SEPOLIA_FACTORY_DEPLOYMENT_MANIFEST"
)

type EVMSweeperSecretConfig struct {
	Network           valueobjects.NetworkID
	RPCURL            string
	SweeperPrivateKey string
}

type EVMSweeperNetworkConfig struct {
	Network           valueobjects.NetworkID
	RPCURL            string
	SweeperPrivateKey string
	FactoryID         int64
	FactoryAddress    string
	CollectorAddress  string
}

type EthereumFactoryDeployDefaults struct {
	Network                valueobjects.NetworkID
	RPCURL                 string
	DeployPrivateKey       string
	Confirmations          int
	CollectorAddress       string
	DeploymentManifestPath string
}

type EthereumFactoryDeploymentManifest struct {
	ContractName              string    `json:"contractName"`
	ChainID                   string    `json:"chainId"`
	ContractAddress           string    `json:"contractAddress"`
	Collector                 string    `json:"collector"`
	VaultCreationCodeHash     string    `json:"vaultCreationCodeHash"`
	DeploymentTransactionHash string    `json:"deploymentTransactionHash"`
	DeployedAt                time.Time `json:"deployedAt"`
}

func LoadEVMSweeperSecretConfigsFromEnv() (map[valueobjects.NetworkID]EVMSweeperSecretConfig, error) {
	configs := make(map[valueobjects.NetworkID]EVMSweeperSecretConfig, 2)

	mainnetConfig, ok := loadEVMSweeperSecretConfigFromEnv(
		valueobjects.NetworkID("mainnet"),
		envEthereumMainnetRPCURL,
		envEthereumMainnetSweeperPrivateKey,
	)
	if ok {
		configs[mainnetConfig.Network] = mainnetConfig
	}

	sepoliaConfig, ok := loadEVMSweeperSecretConfigFromEnv(
		valueobjects.NetworkID("sepolia"),
		envEthereumSepoliaRPCURL,
		envEthereumSepoliaSweeperPrivateKey,
	)
	if ok {
		configs[sepoliaConfig.Network] = sepoliaConfig
	}

	return configs, nil
}

func BuildEVMSweeperNetworkConfigs(
	secretConfigs map[valueobjects.NetworkID]EVMSweeperSecretConfig,
	factoryRecords []outport.EVMFactoryRecord,
) (map[valueobjects.NetworkID]EVMSweeperNetworkConfig, error) {
	configs := make(map[valueobjects.NetworkID]EVMSweeperNetworkConfig, len(secretConfigs))
	activeFactoriesByNetwork := make(map[valueobjects.NetworkID]outport.EVMFactoryRecord, len(factoryRecords))
	for _, record := range factoryRecords {
		if record.Status != outport.EVMFactoryStatusActive {
			continue
		}
		activeFactoriesByNetwork[record.Network] = record
	}

	for network, secretConfig := range secretConfigs {
		if strings.TrimSpace(secretConfig.RPCURL) == "" {
			return nil, fmt.Errorf("ethereum %s rpc url is required", network)
		}
		config := EVMSweeperNetworkConfig{
			Network:           network,
			RPCURL:            secretConfig.RPCURL,
			SweeperPrivateKey: secretConfig.SweeperPrivateKey,
		}
		if record, ok := activeFactoriesByNetwork[network]; ok {
			if !isHexAddress(record.FactoryAddress) {
				return nil, fmt.Errorf("active ethereum %s factory address is invalid", network)
			}
			if !isHexAddress(record.CollectorAddress) {
				return nil, fmt.Errorf("active ethereum %s collector address is invalid", network)
			}
			config.FactoryID = record.ID
			config.FactoryAddress = record.FactoryAddress
			config.CollectorAddress = record.CollectorAddress
		}
		configs[network] = config
	}
	return configs, nil
}

func ConfiguredEVMSweeperNetworks(configs map[valueobjects.NetworkID]EVMSweeperNetworkConfig) []string {
	networks := make([]string, 0, len(configs))
	for network := range configs {
		networks = append(networks, string(network))
	}
	slices.Sort(networks)
	return networks
}

func LoadEthereumFactoryDeploymentManifest(path string) (EthereumFactoryDeploymentManifest, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return EthereumFactoryDeploymentManifest{}, fmt.Errorf("deployment manifest path is required")
	}

	content, err := os.ReadFile(filepath.Clean(trimmedPath))
	if err != nil {
		return EthereumFactoryDeploymentManifest{}, fmt.Errorf("deployment manifest could not be read: %w", err)
	}

	var manifest EthereumFactoryDeploymentManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return EthereumFactoryDeploymentManifest{}, fmt.Errorf("deployment manifest is not valid json: %w", err)
	}

	manifest.ChainID = strings.TrimSpace(manifest.ChainID)
	manifest.ContractAddress = strings.TrimSpace(manifest.ContractAddress)
	manifest.Collector = strings.TrimSpace(manifest.Collector)
	manifest.VaultCreationCodeHash = strings.TrimSpace(manifest.VaultCreationCodeHash)
	manifest.DeploymentTransactionHash = strings.TrimSpace(manifest.DeploymentTransactionHash)
	if !manifest.DeployedAt.IsZero() {
		manifest.DeployedAt = manifest.DeployedAt.UTC()
	}
	if manifest.ChainID == "" {
		return EthereumFactoryDeploymentManifest{}, fmt.Errorf("deployment manifest is missing chainId")
	}
	if manifest.ContractAddress == "" {
		return EthereumFactoryDeploymentManifest{}, fmt.Errorf("deployment manifest is missing contractAddress")
	}
	if manifest.Collector == "" {
		return EthereumFactoryDeploymentManifest{}, fmt.Errorf("deployment manifest is missing collector")
	}
	if manifest.VaultCreationCodeHash == "" {
		return EthereumFactoryDeploymentManifest{}, fmt.Errorf("deployment manifest is missing vaultCreationCodeHash")
	}

	return manifest, nil
}

func ParseEthereumNetworkFromChainID(chainID string) (valueobjects.NetworkID, bool) {
	switch strings.TrimSpace(chainID) {
	case "1":
		return valueobjects.NetworkID("mainnet"), true
	case "11155111":
		return valueobjects.NetworkID("sepolia"), true
	default:
		return "", false
	}
}

func LoadEthereumFactoryDeployDefaultsFromEnv(
	network valueobjects.NetworkID,
) (EthereumFactoryDeployDefaults, bool, error) {
	switch network {
	case valueobjects.NetworkID("mainnet"):
		return ethereumFactoryDeployDefaultsFromEnv(
			network,
			envEthereumMainnetRPCURL,
			envEthereumMainnetDeployPrivateKey,
			envEthereumMainnetDeployConfirmations,
			envEthereumMainnetCollectorAddress,
			envEthereumMainnetFactoryManifest,
		)
	case valueobjects.NetworkID("sepolia"):
		return ethereumFactoryDeployDefaultsFromEnv(
			network,
			envEthereumSepoliaRPCURL,
			envEthereumSepoliaDeployPrivateKey,
			envEthereumSepoliaDeployConfirmations,
			envEthereumSepoliaCollectorAddress,
			envEthereumSepoliaFactoryManifest,
		)
	default:
		return EthereumFactoryDeployDefaults{}, false, nil
	}
}

func loadEVMSweeperSecretConfigFromEnv(
	network valueobjects.NetworkID,
	rpcURLKey string,
	privateKeyKey string,
) (EVMSweeperSecretConfig, bool) {
	rpcURL := strings.TrimSpace(os.Getenv(rpcURLKey))
	sweeperPrivateKey := strings.TrimSpace(os.Getenv(privateKeyKey))
	if countNonEmpty(rpcURL, sweeperPrivateKey) == 0 {
		return EVMSweeperSecretConfig{}, false
	}
	return EVMSweeperSecretConfig{
		Network:           network,
		RPCURL:            rpcURL,
		SweeperPrivateKey: sweeperPrivateKey,
	}, true
}

func ethereumFactoryDeployDefaultsFromEnv(
	network valueobjects.NetworkID,
	rpcURLKey string,
	privateKeyKey string,
	confirmationsKey string,
	collectorKey string,
	manifestKey string,
) (EthereumFactoryDeployDefaults, bool, error) {
	defaults := EthereumFactoryDeployDefaults{
		Network:                network,
		RPCURL:                 strings.TrimSpace(os.Getenv(rpcURLKey)),
		DeployPrivateKey:       strings.TrimSpace(os.Getenv(privateKeyKey)),
		CollectorAddress:       strings.TrimSpace(os.Getenv(collectorKey)),
		DeploymentManifestPath: strings.TrimSpace(os.Getenv(manifestKey)),
	}
	if raw := strings.TrimSpace(os.Getenv(confirmationsKey)); raw != "" {
		confirmations, err := parseEVMConfirmations(raw)
		if err != nil {
			return EthereumFactoryDeployDefaults{}, false, err
		}
		defaults.Confirmations = confirmations
	} else {
		defaults.Confirmations = 1
	}
	if countNonEmpty(
		defaults.RPCURL,
		defaults.DeployPrivateKey,
		defaults.CollectorAddress,
		defaults.DeploymentManifestPath,
	) == 0 {
		return EthereumFactoryDeployDefaults{}, false, nil
	}
	return defaults, true, nil
}

func parseEVMConfirmations(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, fmt.Errorf("confirmations value is required")
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("confirmations value is invalid: %w", err)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("confirmations must be greater than or equal to zero")
	}
	return parsed, nil
}

func countNonEmpty(values ...string) int {
	count := 0
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			count++
		}
	}
	return count
}

func isHexAddress(value string) bool {
	if len(value) != 42 || !strings.HasPrefix(value, "0x") {
		return false
	}
	for i := 2; i < len(value); i++ {
		switch char := value[i]; {
		case char >= '0' && char <= '9':
		case char >= 'a' && char <= 'f':
		case char >= 'A' && char <= 'F':
		default:
			return false
		}
	}
	return true
}
