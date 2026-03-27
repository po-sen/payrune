package bootstrap

import (
	"context"
	"encoding/json"
	"time"

	scheduleradapter "payrune/internal/adapters/inbound/scheduler"
	"payrune/internal/adapters/outbound/bitcoin"
	blockchainadapter "payrune/internal/adapters/outbound/blockchain"
	"payrune/internal/adapters/outbound/ethereum"
	cloudflarepostgresadapter "payrune/internal/adapters/outbound/persistence/cloudflarepostgres"
	"payrune/internal/adapters/outbound/system"
	outport "payrune/internal/application/ports/outbound"
	"payrune/internal/application/usecases"
	"payrune/internal/domain/policies"
	"payrune/internal/domain/valueobjects"
	cloudflarepostgresinfra "payrune/internal/infrastructure/cloudflarepostgres"
)

const (
	cloudflarePollerDefaultBatchSize          = 2
	cloudflarePollerDefaultRescheduleInterval = 10 * time.Minute
	cloudflarePollerDefaultClaimTTL           = 30 * time.Second
)

type pollerWorkerRequestEnvelope struct {
	Env              map[string]string `json:"env"`
	PostgresBridgeID string            `json:"postgresBridgeId"`
	BitcoinBridgeID  string            `json:"bitcoinBridgeId"`
	ScheduledTime    string            `json:"scheduledTime"`
	Cron             string            `json:"cron"`
}

type pollerWorkerResponseEnvelope struct {
	Output scheduleradapter.PollerResponse `json:"output"`
}

func HandleCloudflarePollerRequestJSON(ctx context.Context, payload string) (string, error) {
	var envelope pollerWorkerRequestEnvelope
	if err := json.Unmarshal([]byte(payload), &envelope); err != nil {
		return "", err
	}

	output, err := handleCloudflarePollerRequest(ctx, envelope)
	if err != nil {
		return "", err
	}

	encoded, err := json.Marshal(pollerWorkerResponseEnvelope{Output: output})
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func handleCloudflarePollerRequest(
	ctx context.Context,
	envelope pollerWorkerRequestEnvelope,
) (scheduleradapter.PollerResponse, error) {
	handler, request, err := buildCloudflarePollerRuntime(
		envelope.Env,
		envelope.PostgresBridgeID,
		envelope.BitcoinBridgeID,
	)
	if err != nil {
		return scheduleradapter.PollerResponse{}, err
	}

	return handler.Handle(ctx, request)
}

func buildCloudflarePollerRuntime(
	env map[string]string,
	postgresBridgeID string,
	bitcoinBridgeID string,
) (*scheduleradapter.PollerHandler, scheduleradapter.PollerRequest, error) {
	request, err := buildCloudflarePollerRequest(env)
	if err != nil {
		return nil, scheduleradapter.PollerRequest{}, err
	}

	clock := system.NewClock()
	unitOfWork := cloudflarepostgresadapter.NewUnitOfWork(postgresBridgeID, cloudflarepostgresinfra.NewJSBridge())
	chainObservers := make(map[valueobjects.ChainID]outport.ChainReceiptObserver, 2)
	if request.Chain == "" || request.Chain == valueobjects.ChainIDBitcoin {
		chainObservers[valueobjects.ChainIDBitcoin] = bitcoin.NewCloudflareBitcoinEsploraReceiptObserver(
			bitcoinBridgeID,
			bitcoin.NewCloudflareEsploraBridge(),
		)
	}
	if request.Chain == "" || request.Chain == valueobjects.ChainIDEthereum {
		if ethereumConfigs := loadEthereumRPCConfigsFromLookup(func(key string) string {
			return env[key]
		}); len(ethereumConfigs) > 0 {
			ethereumObserver, err := ethereum.NewEthereumRPCReceiptObserver(ethereumConfigs)
			if err != nil {
				return nil, scheduleradapter.PollerRequest{}, err
			}
			chainObservers[valueobjects.ChainIDEthereum] = ethereumObserver
		}
	}
	receiptObserver, err := blockchainadapter.NewMultiChainReceiptObserver(chainObservers)
	if err != nil {
		return nil, scheduleradapter.PollerRequest{}, err
	}

	useCase := usecases.NewRunReceiptPollingCycleUseCase(
		unitOfWork,
		receiptObserver,
		clock,
		policies.NewPaymentReceiptTrackingLifecyclePolicy(),
	)

	handler := scheduleradapter.NewPollerHandler(scheduleradapter.PollerDependencies{
		RunReceiptPollingCycleUseCase: useCase,
	})
	return handler, request, nil
}

func buildCloudflarePollerRequest(env map[string]string) (scheduleradapter.PollerRequest, error) {
	dispatchConfig, err := loadPollerDispatchConfigFromLookup(func(key string) string {
		return env[key]
	}, pollerDispatchDefaults{
		RescheduleInterval: cloudflarePollerDefaultRescheduleInterval,
		BatchSize:          cloudflarePollerDefaultBatchSize,
		ClaimTTL:           cloudflarePollerDefaultClaimTTL,
	})
	if err != nil {
		return scheduleradapter.PollerRequest{}, err
	}

	return scheduleradapter.PollerRequest{
		BatchSize:          dispatchConfig.BatchSize,
		RescheduleInterval: dispatchConfig.RescheduleInterval,
		ClaimTTL:           dispatchConfig.ClaimTTL,
		Chain:              dispatchConfig.Chain,
		Network:            dispatchConfig.Network,
	}, nil
}
