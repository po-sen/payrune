package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	ethereumadapter "payrune/internal/adapters/outbound/ethereum"
	postgresadapter "payrune/internal/adapters/outbound/persistence/postgres"
	"payrune/internal/application/dto"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/valueobjects"
	"payrune/internal/infrastructure/di"
	postgresdriver "payrune/internal/infrastructure/drivers/postgres"
)

var (
	loadEVMSweeperSecretConfigs = di.LoadEVMSweeperSecretConfigsFromEnv
	buildEVMSweeperRuntime      = di.BuildEVMSweeperNetworkConfigs
	openEVMSweeperDB            = postgresdriver.OpenFromEnv
	loadActiveEVMFactories      = func(ctx context.Context, db *sql.DB) ([]outport.EVMFactoryRecord, error) {
		return postgresadapter.NewEVMFactoryStore(db).ListActive(ctx)
	}
	executeEVMSweeper = runEVMSweeperUseCase
)

type EVMSweeperConfig struct {
	Network           string
	AssetCode         string
	PaymentAddressIDs []int64
	BeforeIssuedAt    time.Time
	BatchSize         int
	DryRun            bool
}

func RunEVMSweeper(ctx context.Context, config EVMSweeperConfig) error {
	secretConfigs, err := loadEVMSweeperSecretConfigs()
	if err != nil {
		return err
	}
	db, err := openEVMSweeperDB()
	if err != nil {
		return err
	}
	if db != nil {
		defer func() {
			_ = db.Close()
		}()
	}

	factoryRecords, err := loadActiveEVMFactories(ctx, db)
	if err != nil {
		return err
	}

	runtimeConfigs, err := buildEVMSweeperRuntime(secretConfigs, factoryRecords)
	if err != nil {
		return err
	}

	configuredNetworks := di.ConfiguredEVMSweeperNetworks(runtimeConfigs)
	log.Printf(
		"evm sweeper invoked dry_run=%t network=%q asset_code=%q payment_address_ids=%v before_issued_at=%q batch_size=%d configured_networks=%v",
		config.DryRun,
		config.Network,
		config.AssetCode,
		config.PaymentAddressIDs,
		formatBootstrapTime(config.BeforeIssuedAt),
		config.BatchSize,
		configuredNetworks,
	)

	if config.Network != "" {
		if _, ok := runtimeConfigs[valueobjects.NetworkID(config.Network)]; !ok {
			return errors.New("requested ethereum sweeper network is not configured")
		}
	}
	if len(runtimeConfigs) == 0 {
		if config.DryRun {
			log.Printf("evm sweeper dry-run skipped because no ethereum networks are configured")
			return nil
		}
		return errors.New("at least one ethereum sweeper network config is required")
	}

	if config.DryRun {
		return executeEVMSweeper(ctx, db, runtimeConfigs, config, true)
	}
	if err := validateEVMSweeperExecuteConfig(runtimeConfigs, config.Network); err != nil {
		return err
	}

	return executeEVMSweeper(ctx, db, runtimeConfigs, config, false)
}

func formatBootstrapTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func validateEVMSweeperExecuteConfig(
	runtimeConfigs map[valueobjects.NetworkID]di.EVMSweeperNetworkConfig,
	selectedNetwork string,
) error {
	if selectedNetwork != "" {
		config, ok := runtimeConfigs[valueobjects.NetworkID(selectedNetwork)]
		if !ok {
			return errors.New("requested ethereum sweeper network is not configured")
		}
		if config.SweeperPrivateKey == "" {
			return errors.New("requested ethereum sweeper network is missing sweeper private key")
		}
		return nil
	}

	for _, network := range di.ConfiguredEVMSweeperNetworks(runtimeConfigs) {
		if runtimeConfigs[valueobjects.NetworkID(network)].SweeperPrivateKey == "" {
			return errors.New("configured ethereum sweeper network is missing sweeper private key")
		}
	}
	return nil
}

func runEVMSweeperUseCase(
	ctx context.Context,
	db *sql.DB,
	runtimeConfigs map[valueobjects.NetworkID]di.EVMSweeperNetworkConfig,
	config EVMSweeperConfig,
	dryRun bool,
) error {
	vaultStore := postgresadapter.NewEVMPaymentVaultStore(db)
	var executor outport.EVMSweepExecutor
	if !dryRun {
		executor = ethereumadapter.NewSweepExecutor()
	}

	useCase := usecases.NewRunEVMSweepUseCase(
		vaultStore,
		executor,
		mapEVMSweeperRuntimes(runtimeConfigs),
	)
	output, err := useCase.Execute(ctx, dto.RunEVMSweepInput{
		Network:           valueobjects.NetworkID(config.Network),
		AssetCode:         config.AssetCode,
		PaymentAddressIDs: config.PaymentAddressIDs,
		BeforeIssuedAt:    config.BeforeIssuedAt,
		BatchSize:         config.BatchSize,
		DryRun:            dryRun,
	})
	for _, batch := range output.Batches {
		log.Printf(
			"evm sweeper batch status=%q network=%q factory_address=%q asset_code=%q asset_type=%q token_address=%q payment_address_ids=%v tx_hash=%q error=%q",
			batch.Status,
			batch.Network,
			batch.FactoryAddress,
			batch.AssetCode,
			batch.AssetType,
			batch.TokenAddress,
			batch.PaymentAddressIDs,
			batch.TxHash,
			batch.Error,
		)
	}
	log.Printf(
		"evm sweeper completed dry_run=%t candidate_count=%d batch_count=%d",
		dryRun,
		output.CandidateCount,
		output.BatchCount,
	)
	return err
}

func mapEVMSweeperRuntimes(
	configs map[valueobjects.NetworkID]di.EVMSweeperNetworkConfig,
) map[valueobjects.NetworkID]dto.EVMSweepNetworkRuntime {
	runtimes := make(map[valueobjects.NetworkID]dto.EVMSweepNetworkRuntime, len(configs))
	for network, config := range configs {
		runtimes[network] = dto.EVMSweepNetworkRuntime{
			Network:           network,
			RPCURL:            config.RPCURL,
			SweeperPrivateKey: config.SweeperPrivateKey,
		}
	}
	return runtimes
}
