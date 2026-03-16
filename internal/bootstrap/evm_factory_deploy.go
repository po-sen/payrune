package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	ethereumadapter "payrune/internal/adapters/outbound/ethereum"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	"payrune/internal/adapters/outbound/system"
	"payrune/internal/application/dto"
	inport "payrune/internal/application/ports/inbound"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/valueobjects"
	"payrune/internal/infrastructure/di"
	postgresdriver "payrune/internal/infrastructure/drivers/postgres"
)

type EVMFactoryDeployConfig struct {
	Network                string
	RPCURL                 string
	DeployPrivateKey       string
	CollectorAddress       string
	Confirmations          int
	OutputManifestPath     string
	DeploymentManifestPath string
	ContractsArtifactPath  string
	ReplaceActive          bool
}

func RunEVMFactoryDeploy(ctx context.Context, config EVMFactoryDeployConfig) error {
	db, err := postgresdriver.OpenFromEnv()
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	factoryStore := postgresadapter.NewEVMFactoryStore(db)
	registerUseCase := usecases.NewRegisterEVMFactoryUseCase(
		postgresadapter.NewUnitOfWork(db),
		system.NewClock(),
	)

	if strings.TrimSpace(config.DeploymentManifestPath) != "" {
		return registerEVMFactoryFromManifest(ctx, registerUseCase, config)
	}

	if err := ensureEVMFactoryDeployAllowed(ctx, factoryStore, config.Network, config.ReplaceActive); err != nil {
		return err
	}

	deployer := ethereumadapter.NewFactoryDeployer()
	deployOutput, err := deployer.Deploy(ctx, outport.DeployEVMFactoryInput{
		Network:               valueobjects.NetworkID(strings.TrimSpace(config.Network)),
		RPCURL:                config.RPCURL,
		DeployPrivateKey:      config.DeployPrivateKey,
		CollectorAddress:      config.CollectorAddress,
		Confirmations:         config.Confirmations,
		OutputManifestPath:    config.OutputManifestPath,
		ContractsArtifactPath: config.ContractsArtifactPath,
	})
	if err != nil {
		return err
	}

	response, err := registerUseCase.Execute(ctx, dto.RegisterEVMFactoryInput{
		Network:               valueobjects.NetworkID(strings.TrimSpace(config.Network)),
		FactoryAddress:        deployOutput.Manifest.ContractAddress,
		CollectorAddress:      deployOutput.Manifest.Collector,
		VaultCreationCodeHash: deployOutput.Manifest.VaultCreationCodeHash,
		DeploymentTxHash:      deployOutput.Manifest.DeploymentTransactionHash,
		DeployedAt:            deployOutput.Manifest.DeployedAt,
		AllowReplaceActive:    config.ReplaceActive,
	})
	if err != nil {
		return err
	}

	log.Printf(
		"evm factory deployed and registered network=%q factory_address=%q collector_address=%q vault_creation_code_hash=%q status=%q deployment_tx_hash=%q deployed_at=%q manifest_path=%q replace_active=%t",
		response.Network,
		response.FactoryAddress,
		response.CollectorAddress,
		response.VaultCreationCodeHash,
		response.Status,
		response.DeploymentTxHash,
		formatBootstrapTime(response.DeployedAt),
		deployOutput.OutputManifestPath,
		config.ReplaceActive,
	)
	return nil
}

func registerEVMFactoryFromManifest(
	ctx context.Context,
	registerUseCase inport.RegisterEVMFactoryUseCase,
	config EVMFactoryDeployConfig,
) error {
	manifest, err := di.LoadEthereumFactoryDeploymentManifest(config.DeploymentManifestPath)
	if err != nil {
		return err
	}

	network, ok := di.ParseEthereumNetworkFromChainID(manifest.ChainID)
	if !ok {
		return fmt.Errorf("unsupported ethereum chainId in deployment manifest: %s", manifest.ChainID)
	}
	if strings.TrimSpace(config.Network) != "" &&
		network != valueobjects.NetworkID(strings.TrimSpace(config.Network)) {
		return fmt.Errorf(
			"deployment manifest chainId %s does not match requested network %s",
			manifest.ChainID,
			config.Network,
		)
	}

	response, err := registerUseCase.Execute(ctx, dto.RegisterEVMFactoryInput{
		Network:               network,
		FactoryAddress:        manifest.ContractAddress,
		CollectorAddress:      manifest.Collector,
		VaultCreationCodeHash: manifest.VaultCreationCodeHash,
		DeploymentTxHash:      manifest.DeploymentTransactionHash,
		DeployedAt:            manifest.DeployedAt,
		AllowReplaceActive:    config.ReplaceActive,
	})
	if err != nil {
		return err
	}

	log.Printf(
		"evm factory registered from manifest network=%q factory_address=%q collector_address=%q vault_creation_code_hash=%q status=%q deployment_tx_hash=%q deployed_at=%q manifest_path=%q replace_active=%t",
		response.Network,
		response.FactoryAddress,
		response.CollectorAddress,
		response.VaultCreationCodeHash,
		response.Status,
		response.DeploymentTxHash,
		formatBootstrapTime(response.DeployedAt),
		config.DeploymentManifestPath,
		config.ReplaceActive,
	)
	return nil
}

func ensureEVMFactoryDeployAllowed(
	ctx context.Context,
	factoryStore outport.EVMFactoryStore,
	networkRaw string,
	replaceActive bool,
) error {
	if replaceActive {
		return nil
	}
	if factoryStore == nil {
		return errors.New("evm factory registry store is not configured")
	}

	record, found, err := factoryStore.FindActiveByNetwork(ctx, valueobjects.NetworkID(strings.TrimSpace(networkRaw)))
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	return fmt.Errorf(
		"active evm factory already exists for network %s at %s; rerun with --replace-active to rotate it",
		record.Network,
		record.FactoryAddress,
	)
}
