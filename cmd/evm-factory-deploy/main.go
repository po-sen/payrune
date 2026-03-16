package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"payrune/internal/bootstrap"
	"payrune/internal/domain/valueobjects"
	"payrune/internal/infrastructure/di"
)

const defaultContractsArtifactPath = "deployments/ethereum/build/contracts.json"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := parseEVMFactoryDeployConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("invalid evm factory deploy config: %v", err)
	}

	if err := bootstrap.RunEVMFactoryDeploy(ctx, config); err != nil {
		log.Fatalf("evm factory deploy exited with error: %v", err)
	}
}

func parseEVMFactoryDeployConfig(args []string) (bootstrap.EVMFactoryDeployConfig, error) {
	flagSet := flag.NewFlagSet("evm-factory-deploy", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)

	var (
		networkRaw                string
		rpcURLRaw                 string
		deployPrivateKeyRaw       string
		collectorAddressRaw       string
		outputManifestPathRaw     string
		deploymentManifestPathRaw string
		contractsArtifactPathRaw  string
		confirmationsRaw          int
		replaceActive             bool
	)

	flagSet.StringVar(&networkRaw, "network", "", "target network")
	flagSet.StringVar(&rpcURLRaw, "rpc-url", "", "ethereum rpc url")
	flagSet.StringVar(&deployPrivateKeyRaw, "deploy-private-key", "", "hex deployer private key")
	flagSet.StringVar(&collectorAddressRaw, "collector-address", "", "collector account address")
	flagSet.StringVar(&outputManifestPathRaw, "output-manifest", "", "path to write deployment manifest json")
	flagSet.StringVar(
		&deploymentManifestPathRaw,
		"deployment-manifest",
		"",
		"path to an existing deployment manifest to register without deploying",
	)
	flagSet.StringVar(
		&contractsArtifactPathRaw,
		"contracts-artifact",
		"",
		"path to compiled contracts artifact json",
	)
	flagSet.IntVar(&confirmationsRaw, "confirmations", -1, "required confirmations before registering")
	flagSet.BoolVar(&replaceActive, "replace-active", false, "retire the current active factory for this network")

	if err := flagSet.Parse(args); err != nil {
		return bootstrap.EVMFactoryDeployConfig{}, err
	}

	config := bootstrap.EVMFactoryDeployConfig{
		Network:                strings.TrimSpace(networkRaw),
		RPCURL:                 strings.TrimSpace(rpcURLRaw),
		DeployPrivateKey:       strings.TrimSpace(deployPrivateKeyRaw),
		CollectorAddress:       strings.TrimSpace(collectorAddressRaw),
		OutputManifestPath:     strings.TrimSpace(outputManifestPathRaw),
		DeploymentManifestPath: strings.TrimSpace(deploymentManifestPathRaw),
		ContractsArtifactPath:  strings.TrimSpace(contractsArtifactPathRaw),
		Confirmations:          confirmationsRaw,
		ReplaceActive:          replaceActive,
	}

	if config.DeploymentManifestPath != "" {
		manifest, err := di.LoadEthereumFactoryDeploymentManifest(config.DeploymentManifestPath)
		if err != nil {
			return bootstrap.EVMFactoryDeployConfig{}, err
		}
		network, ok := di.ParseEthereumNetworkFromChainID(manifest.ChainID)
		if !ok {
			return bootstrap.EVMFactoryDeployConfig{}, fmt.Errorf(
				"unsupported ethereum chainId in deployment manifest: %s",
				manifest.ChainID,
			)
		}
		if config.Network == "" {
			config.Network = string(network)
		} else if config.Network != string(network) {
			return bootstrap.EVMFactoryDeployConfig{}, fmt.Errorf(
				"deployment manifest chainId %s does not match requested network %s",
				manifest.ChainID,
				config.Network,
			)
		}
	}

	if config.Network != "" {
		defaults, ok, err := di.LoadEthereumFactoryDeployDefaultsFromEnv(valueobjects.NetworkID(config.Network))
		if err != nil {
			return bootstrap.EVMFactoryDeployConfig{}, err
		}
		if ok {
			if config.RPCURL == "" {
				config.RPCURL = defaults.RPCURL
			}
			if config.DeployPrivateKey == "" {
				config.DeployPrivateKey = defaults.DeployPrivateKey
			}
			if config.CollectorAddress == "" {
				config.CollectorAddress = defaults.CollectorAddress
			}
			if config.OutputManifestPath == "" {
				config.OutputManifestPath = defaults.DeploymentManifestPath
			}
			if config.Confirmations < 0 {
				config.Confirmations = defaults.Confirmations
			}
		}
	}

	if config.ContractsArtifactPath == "" {
		config.ContractsArtifactPath = defaultContractsArtifactPath
	}
	if config.Confirmations < 0 {
		config.Confirmations = 1
	}

	if config.Network == "" {
		return bootstrap.EVMFactoryDeployConfig{}, fmt.Errorf("network is required")
	}
	if !isSupportedEthereumNetwork(config.Network) {
		return bootstrap.EVMFactoryDeployConfig{}, fmt.Errorf("network must be mainnet or sepolia")
	}

	if config.DeploymentManifestPath != "" {
		return config, nil
	}

	if config.RPCURL == "" {
		return bootstrap.EVMFactoryDeployConfig{}, fmt.Errorf("rpc-url is required")
	}
	if config.DeployPrivateKey == "" {
		return bootstrap.EVMFactoryDeployConfig{}, fmt.Errorf("deploy-private-key is required")
	}
	if !isHexAddress(config.CollectorAddress) {
		return bootstrap.EVMFactoryDeployConfig{}, fmt.Errorf("collector-address must be a valid hex address")
	}
	if config.Confirmations < 0 {
		return bootstrap.EVMFactoryDeployConfig{}, fmt.Errorf("confirmations must be greater than or equal to zero")
	}

	return config, nil
}

func isSupportedEthereumNetwork(value string) bool {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "mainnet", "sepolia":
		return true
	default:
		return false
	}
}

func isHexAddress(value string) bool {
	raw := strings.TrimSpace(value)
	if len(raw) != 42 || !strings.HasPrefix(raw, "0x") {
		return false
	}
	for i := 2; i < len(raw); i++ {
		switch char := raw[i]; {
		case char >= '0' && char <= '9':
		case char >= 'a' && char <= 'f':
		case char >= 'A' && char <= 'F':
		default:
			return false
		}
	}
	return true
}
