package bootstrap

import (
	"context"
	"database/sql"
	"testing"

	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/domain/valueobjects"
	"payrune/internal/infrastructure/di"
)

func TestRunEVMSweeperDryRunWithoutConfiguredNetworks(t *testing.T) {
	withMockedEVMSweeperRuntime(
		t,
		map[valueobjects.NetworkID]di.EVMSweeperSecretConfig{},
		[]outport.EVMFactoryRecord{},
	)

	err := RunEVMSweeper(context.Background(), EVMSweeperConfig{DryRun: true})
	if err != nil {
		t.Fatalf("RunEVMSweeper returned error: %v", err)
	}
}

func TestRunEVMSweeperRejectsRequestedNetworkWithoutRuntimeConfig(t *testing.T) {
	withMockedEVMSweeperRuntime(
		t,
		map[valueobjects.NetworkID]di.EVMSweeperSecretConfig{},
		[]outport.EVMFactoryRecord{},
	)

	err := RunEVMSweeper(context.Background(), EVMSweeperConfig{
		Network: "sepolia",
		DryRun:  true,
	})
	if err == nil {
		t.Fatal("expected missing network config error")
	}
}

func TestRunEVMSweeperExecuteModeRequiresAtLeastOneConfiguredNetwork(t *testing.T) {
	withMockedEVMSweeperRuntime(
		t,
		map[valueobjects.NetworkID]di.EVMSweeperSecretConfig{},
		[]outport.EVMFactoryRecord{},
	)

	err := RunEVMSweeper(context.Background(), EVMSweeperConfig{DryRun: false})
	if err == nil {
		t.Fatal("expected missing network config error")
	}
}

func TestRunEVMSweeperDryRunAllowsConfiguredNetworkWithoutPrivateKey(t *testing.T) {
	withMockedEVMSweeperRuntime(
		t,
		map[valueobjects.NetworkID]di.EVMSweeperSecretConfig{
			valueobjects.NetworkID("sepolia"): {
				Network: valueobjects.NetworkID("sepolia"),
				RPCURL:  "https://sepolia.example",
			},
		},
		[]outport.EVMFactoryRecord{
			{
				Network:          valueobjects.NetworkID("sepolia"),
				FactoryAddress:   "0x1111111111111111111111111111111111111111",
				CollectorAddress: "0x2222222222222222222222222222222222222222",
				Status:           outport.EVMFactoryStatusActive,
			},
		},
	)

	err := RunEVMSweeper(context.Background(), EVMSweeperConfig{
		Network: "sepolia",
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("RunEVMSweeper returned error: %v", err)
	}
}

func TestRunEVMSweeperExecuteModeRejectsMissingPrivateKey(t *testing.T) {
	withMockedEVMSweeperRuntime(
		t,
		map[valueobjects.NetworkID]di.EVMSweeperSecretConfig{
			valueobjects.NetworkID("sepolia"): {
				Network: valueobjects.NetworkID("sepolia"),
				RPCURL:  "https://sepolia.example",
			},
		},
		[]outport.EVMFactoryRecord{
			{
				Network:          valueobjects.NetworkID("sepolia"),
				FactoryAddress:   "0x1111111111111111111111111111111111111111",
				CollectorAddress: "0x2222222222222222222222222222222222222222",
				Status:           outport.EVMFactoryStatusActive,
			},
		},
	)

	err := RunEVMSweeper(context.Background(), EVMSweeperConfig{
		Network: "sepolia",
		DryRun:  false,
	})
	if err == nil {
		t.Fatal("expected missing private key error")
	}
}

func withMockedEVMSweeperRuntime(
	t *testing.T,
	secretConfigs map[valueobjects.NetworkID]di.EVMSweeperSecretConfig,
	factoryRecords []outport.EVMFactoryRecord,
) {
	t.Helper()

	previousLoadSecrets := loadEVMSweeperSecretConfigs
	previousBuildRuntime := buildEVMSweeperRuntime
	previousOpenDB := openEVMSweeperDB
	previousLoadFactories := loadActiveEVMFactories
	previousExecute := executeEVMSweeper
	t.Cleanup(func() {
		loadEVMSweeperSecretConfigs = previousLoadSecrets
		buildEVMSweeperRuntime = previousBuildRuntime
		openEVMSweeperDB = previousOpenDB
		loadActiveEVMFactories = previousLoadFactories
		executeEVMSweeper = previousExecute
	})

	loadEVMSweeperSecretConfigs = func() (map[valueobjects.NetworkID]di.EVMSweeperSecretConfig, error) {
		return secretConfigs, nil
	}
	buildEVMSweeperRuntime = di.BuildEVMSweeperNetworkConfigs
	openEVMSweeperDB = func() (*sql.DB, error) {
		return nil, nil
	}
	loadActiveEVMFactories = func(context.Context, *sql.DB) ([]outport.EVMFactoryRecord, error) {
		return factoryRecords, nil
	}
	executeEVMSweeper = func(context.Context, *sql.DB, map[valueobjects.NetworkID]di.EVMSweeperNetworkConfig, EVMSweeperConfig, bool) error {
		return nil
	}
}
